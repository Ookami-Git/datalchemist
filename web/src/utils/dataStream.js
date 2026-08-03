import axios from 'axios';

// Le backend expose les données via un flux SSE (`<url>/stream`) qui émet des
// événements `progress` pendant le chargement des sources puis un événement
// `result` avec les données complètes. Si le flux est indisponible (proxy qui
// tamponne, backend plus ancien, EventSource absent), on retombe sur l'appel
// JSON classique : le chargement fonctionne, sans suivi détaillé.

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

export function fetchDataWithProgress({ url, params = {}, onProgress = null }) {
  const queryString = buildQueryString(params);
  const streamUrl = `${url}/stream${queryString ? `?${queryString}` : ''}`;

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

    source.addEventListener('result', (event) => {
      if (canceled || settled) return;
      settled = true;
      closeStream();
      try {
        resolve(JSON.parse(event.data));
      } catch (error) {
        reject(error);
      }
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
