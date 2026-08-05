import axios from 'axios';
import { createDataCollector } from './progressiveData.js';

// Le backend expose les données via un flux SSE (`<url>/stream`) qui émet le plan
// de chargement (`plan`), la valeur de chaque source dès qu'elle est prête
// (`source`), l'avancement (`progress`), puis la fin du chargement (`complete`).
// Si le flux est indisponible (proxy qui tamponne, EventSource absent), on
// retombe sur l'appel JSON classique : le chargement fonctionne, sans suivi
// détaillé ni affichage progressif.

function buildQueryString(params = {}) {
  const searchParams = new URLSearchParams();

  Object.entries(params || {}).forEach(([key, value]) => {
    if (value === undefined || value === null) return;
    if (Array.isArray(value)) {
      value.forEach((entry) => {
        if (entry !== undefined && entry !== null) searchParams.append(key, entry);
      });
      return;
    }
    searchParams.append(key, value);
  });

  return searchParams.toString();
}

// onItemsReady est appelé avec la liste des objets dont toutes les sources sont
// chargées, et un instantané figé des données à leur passer.
export function fetchDataWithProgress({ url, params = {}, onProgress = null, onItemsReady = null }) {
  const queryString = buildQueryString(params);
  const streamUrl = `${url}/stream${queryString ? `?${queryString}` : ''}`;

  const collector = createDataCollector();
  let source = null;
  let settled = false;
  let canceled = false;
  const controller = new AbortController();

  const closeStream = () => {
    if (!source) return;
    // Obligatoire : sans close(), EventSource se reconnecte automatiquement et
    // le backend recalculerait toutes les sources.
    source.close();
    source = null;
  };

  const promise = new Promise((resolve, reject) => {
    const fallbackToJson = () => {
      if (settled || canceled) return;
      closeStream();
      axios.get(url, { params, signal: controller.signal })
        .then((response) => {
          if (canceled) return;
          settled = true;
          resolve(response.data);
        })
        .catch((error) => {
          if (canceled) return;
          settled = true;
          reject(error);
        });
    };

    if (typeof window === 'undefined' || typeof window.EventSource !== 'function') {
      fallbackToJson();
      return;
    }

    try {
      source = new EventSource(streamUrl, { withCredentials: true });
    } catch {
      fallbackToJson();
      return;
    }

    source.addEventListener('progress', (event) => {
      if (canceled || settled || !onProgress) return;
      try {
        onProgress(JSON.parse(event.data));
      } catch {
        // Un événement de progression illisible ne doit pas casser le chargement.
      }
    });

    const notifyReady = (itemids) => {
      if (!onItemsReady || itemids.length === 0) return;
      onItemsReady(itemids, collector.snapshot());
    };

    // Plan de chargement : le frontend sait dès maintenant quelles sources
    // chaque objet attend, et affiche d'emblée ceux qui n'en ont aucune.
    source.addEventListener('plan', (event) => {
      if (canceled || settled) return;
      try {
        notifyReady(collector.setPlan(JSON.parse(event.data)));
      } catch {
        // Un plan illisible ne doit pas casser le chargement : le résultat
        // complet reste servi à la fin du flux.
      }
    });

    source.addEventListener('source', (event) => {
      if (canceled || settled) return;
      try {
        notifyReady(collector.addSource(JSON.parse(event.data)));
      } catch {
        // Idem : une source illisible n'interrompt pas le flux.
      }
    });

    source.addEventListener('complete', () => {
      if (canceled || settled) return;
      settled = true;
      closeStream();
      resolve(collector.data);
    });

    // Événement métier volontairement nommé "failure" : un événement SSE nommé
    // "error" serait confondu avec l'erreur de transport d'EventSource.
    source.addEventListener('failure', (event) => {
      if (canceled || settled) return;
      settled = true;
      closeStream();
      let message = 'Data stream error';
      try {
        message = JSON.parse(event.data).message || message;
      } catch {
        // Message par défaut.
      }
      reject(new Error(message));
    });

    // Erreur de transport : flux coupé ou refusé, on tente l'appel classique.
    source.addEventListener('error', () => {
      if (canceled || settled) return;
      fallbackToJson();
    });
  });

  return {
    promise,
    cancel() {
      canceled = true;
      closeStream();
      controller.abort();
    }
  };
}
