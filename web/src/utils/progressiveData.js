// Accumulation des données d'une vue au fil du flux SSE.
//
// Le backend annonce d'abord le plan de chargement (quelles sources chaque objet
// attend), puis publie chaque source dès qu'elle est prête. Un objet devient
// affichable quand toutes ses sources sont arrivées : on lui fabrique alors un
// instantané figé des données, que les sources suivantes ne modifieront plus.

// createDataCollector construit l'accumulateur d'un chargement.
export function createDataCollector() {
  const data = { sn: {}, sid: {}, get: {} };
  const loaded = new Set();
  let items = null;

  // snapshot fige l'état courant : chaque objet rend avec sa propre référence,
  // sinon les arrivées suivantes le feraient rendre à nouveau.
  const snapshot = () => ({ ...data, sn: { ...data.sn }, sid: { ...data.sid } });

  return {
    data,

    // setPlan enregistre les sources attendues par objet et les variables d'URL.
    // Retourne les objets déjà prêts : ceux sans aucune source le sont d'emblée.
    setPlan(plan = {}) {
      items = plan?.plan?.items || plan?.items || {};
      if (plan?.get) data.get = plan.get;
      return this.readyItems();
    },

    // addSource enregistre la valeur d'une source et retourne les objets que
    // cette arrivée rend affichables.
    addSource({ name, id, value } = {}) {
      if (!name) return [];
      data.sn[name] = value;
      if (id) data.sid[`s${id}`] = value;
      loaded.add(name);
      return this.readyItems();
    },

    // readyItems liste les objets dont toutes les sources sont chargées.
    readyItems() {
      if (!items) return [];
      return Object.entries(items)
        .filter(([, sources]) => (sources || []).every((name) => loaded.has(name)))
        .map(([itemid]) => itemid);
    },

    snapshot
  };
}

// mergeSnapshots ajoute un instantané aux objets qui n'en ont pas encore, sans
// toucher aux instantanés déjà distribués. Retourne l'objet d'origine si rien ne
// change : le rendu des objets déjà affichés n'est pas relancé.
export function mergeSnapshots(current, itemids, snapshot) {
  const missing = itemids.filter((itemid) => !current[itemid]);
  if (missing.length === 0) return current;

  const next = { ...current };
  missing.forEach((itemid) => {
    next[itemid] = snapshot;
  });
  return next;
}
