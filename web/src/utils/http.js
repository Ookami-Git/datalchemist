import axios from 'axios';

// Sur un lien dégradé (VPN à faible MTU, perte de paquets), une réponse
// volumineuse peut être tronquée en cours de transfert : le navigateur signale
// alors une erreur de transport, sans réponse HTTP, et l'UI affiche
// « Unknown - Network Error ». La compression côté serveur réduit la fréquence
// du problème sans l'éliminer ; rejouer la requête récupère le reste.

// La grande majorité des routes ne fait que lire la base locale et répond en
// quelques dizaines de millisecondes. Trente secondes laissent donc largement
// la place à un lien lent, tout en transformant une connexion morte en échec
// exploitable — donc en nouvelle tentative — au lieu d'un chargement infini.
export const DEFAULT_TIMEOUT = 30_000;

// Les routes /api/data déclenchent le chargement des sources : URL distantes,
// bases de données, scripts. Le serveur ne leur impose aucune limite de durée,
// et une vue lourde peut légitimement prendre plusieurs minutes. Ce délai n'est
// donc pas un budget de performance mais un garde-fou contre une connexion
// morte : il doit rester hors d'atteinte d'un chargement réel.
export const DATA_TIMEOUT = 600_000;

const DATA_ROUTE = /\/api\/data\//;

// Trois tentatives au total. Au-delà, l'échec est probablement durable, et
// rejouer une route /api/data coûte au serveur un rechargement complet des
// sources.
const MAX_ATTEMPTS = 3;
const BASE_BACKOFF = 300;

// Seules les méthodes sans effet de bord sont rejouables. Rejouer un POST
// créerait un objet en double ou réappliquerait un import : l'application
// utilise POST pour ses écritures, y compris les mises à jour.
const SAFE_METHODS = new Set(['get', 'head', 'options']);

// Un 500 vient de l'application et se reproduira à l'identique ; ces trois-là
// signalent un intermédiaire momentanément indisponible.
const RETRYABLE_STATUS = new Set([502, 503, 504]);

// Le compteur voyage dans la config de la requête, seul objet qui traverse une
// nouvelle tentative. Il doit porter une clé de type chaîne : mergeConfig
// d'axios recopie les propriétés énumérables nommées, mais laisse tomber les
// clés symboliques — un Symbol ici ferait perdre le compteur à chaque
// tentative, donc boucler sans fin sur une requête durablement en échec.
const ATTEMPT = '__retryAttempt';

// backoffDelay étale les tentatives pour ne pas retomber sur la même rafale de
// pertes, avec une part d'aléatoire pour désynchroniser les requêtes parallèles
// d'une même vue.
function backoffDelay(attempt) {
  const base = BASE_BACKOFF * 2 ** attempt;
  return base * (0.75 + Math.random() * 0.5);
}

// wait respecte l'annulation : un utilisateur qui change de vue pendant
// l'attente ne doit pas déclencher la tentative suivante.
function wait(duration, signal) {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(signal.reason ?? new Error('canceled'));
      return;
    }

    const timer = setTimeout(() => {
      signal?.removeEventListener('abort', onAbort);
      resolve();
    }, duration);

    function onAbort() {
      clearTimeout(timer);
      reject(signal.reason ?? new Error('canceled'));
    }

    signal?.addEventListener('abort', onAbort, { once: true });
  });
}

function isRetryable(error) {
  // Une annulation volontaire (changement de vue, composant démonté) n'est pas
  // un échec réseau.
  if (axios.isCancel(error) || error.code === 'ERR_CANCELED') return false;

  const method = (error.config?.method ?? 'get').toLowerCase();
  if (!SAFE_METHODS.has(method)) return false;

  // Pas de réponse du tout : connexion coupée ou délai dépassé. C'est le cas
  // que voit le navigateur quand un transfert est interrompu (XHR ne rend
  // alors aucune réponse) — précisément ce que la compression ne suffit pas à
  // faire disparaître.
  if (!error.response) return true;

  // Une erreur qui porte pourtant une réponse réussie signale un corps arrivé
  // incomplet : le serveur avait bien répondu 200, mais le transfert s'est
  // arrêté avant la fin annoncée par Content-Length. C'est la forme que prend
  // la troncature quand axios n'utilise pas XHR (adaptateur Node, fetch), et
  // il ne faut surtout pas la confondre avec un refus applicatif : sans cette
  // branche, la troncature n'était pas rejouée du tout.
  const status = error.response.status;
  if (status >= 200 && status < 300) return true;

  return RETRYABLE_STATUS.has(status);
}

// installHttpResilience pose un délai d'expiration par défaut et rejoue les
// requêtes sûres qui échouent en transport. À appeler une seule fois au
// démarrage : l'application partage l'instance axios par défaut.
export function installHttpResilience(client = axios) {
  client.defaults.timeout = DEFAULT_TIMEOUT;

  client.interceptors.request.use((config) => {
    // Le délai par défaut d'axios vaut 0, jamais `undefined` : c'est la
    // comparaison avec la valeur par défaut du client, et non un test de
    // présence, qui distingue un appel muet d'un appel qui a choisi son délai.
    // Celui qui a choisi garde le sien, y compris 0 pour ne jamais expirer.
    if (config.timeout === client.defaults.timeout && DATA_ROUTE.test(config.url ?? '')) {
      config.timeout = DATA_TIMEOUT;
    }
    return config;
  });

  client.interceptors.response.use(null, async (error) => {
    const config = error.config;
    if (!config) throw error;

    const attempt = config[ATTEMPT] ?? 0;
    if (attempt >= MAX_ATTEMPTS - 1 || !isRetryable(error)) throw error;

    config[ATTEMPT] = attempt + 1;
    await wait(backoffDelay(attempt), config.signal);

    // Repasser par le client relance la requête à travers les intercepteurs :
    // le compteur porté par la config borne la récursion.
    return client(config);
  });

  return client;
}
