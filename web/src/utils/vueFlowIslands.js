import { createApp } from 'vue';
import { parseFlowConfig } from './vueFlowConfig.js';

/**
 * Îlots Vue Flow dans le HTML rendu par Nunjucks.
 *
 * `v-html` fait un innerHTML : il ne compile pas de composant Vue. Mermaid s'en sort
 * parce qu'il post-traite le DOM, Vue Flow non. On monte donc une mini-app Vue dans
 * chaque conteneur `[data-vueflow]` laissé par le template.
 *
 * Contrat côté template :
 *   <div data-vueflow>{ "nodes": [...], "edges": [...], "height": "420px" }</div>
 */

// Vue Flow pèse lourd et la plupart des items n'en ont pas besoin : le chunk
// n'est chargé qu'au premier îlot rencontré, puis mémorisé.
let islandComponentPromise = null;

function loadIslandComponent() {
  if (!islandComponentPromise) {
    islandComponentPromise = import('@/components/view/items/VueFlowIsland.vue')
      .then((module) => module.default)
      .catch((err) => {
        islandComponentPromise = null; // autorise une nouvelle tentative
        throw err;
      });
  }
  return islandComponentPromise;
}

/**
 * Monte un îlot Vue Flow dans chaque `[data-vueflow]` présent sous `root`.
 *
 * @param {HTMLElement|null} root – racine du HTML rendu (itemRoot)
 * @param {{ theme?: 'light'|'dark', provides?: Record<string, unknown> }} [options]
 * @returns {Promise<Array<{ app: import('vue').App, el: HTMLElement }>>} apps à démonter plus tard
 */
export async function mountVueFlowIslands(root, { theme = 'light', provides = {} } = {}) {
  if (!root) return [];

  const containers = root.querySelectorAll('[data-vueflow]:not([data-vueflow-mounted])');
  if (!containers.length) return [];

  let VueFlowIsland;
  try {
    VueFlowIsland = await loadIslandComponent();
  } catch (err) {
    console.error('Error loading Vue Flow:', err);
    containers.forEach((el) => {
      el.innerHTML = `<div class="text-danger">Vue Flow: failed to load (${err.message})</div>`;
      el.setAttribute('data-vueflow-mounted', 'true');
    });
    return [];
  }

  const apps = [];

  containers.forEach((el, index) => {
    // Le navigateur a déjà décodé les entités HTML de l'autoescape Nunjucks en
    // parsant le innerHTML, donc `| dump` suffit côté template.
    const raw = el.dataset.vueflowConfig ?? el.textContent ?? '';
    const { config, error } = parseFlowConfig(raw, { fallbackId: `dc-flow-${index}` });

    if (error) {
      el.innerHTML = `<div class="text-danger">Vue Flow: ${error}</div>`;
      el.setAttribute('data-vueflow-mounted', 'true');
      return;
    }

    el.textContent = '';
    const host = document.createElement('div');
    el.appendChild(host);
    el.setAttribute('data-vueflow-mounted', 'true');

    const app = createApp(VueFlowIsland, { config, theme, hostElement: el });
    Object.entries(provides).forEach(([key, value]) => app.provide(key, value));

    try {
      app.mount(host);
      apps.push({ app, el });
    } catch (err) {
      console.error('Error mounting Vue Flow island:', err);
      el.innerHTML = `<div class="text-danger">Vue Flow: mount failed (${err.message})</div>`;
    }
  });

  return apps;
}

/**
 * Démonte les îlots et vide le tableau passé en argument.
 *
 * @param {Array<{ app: import('vue').App, el: HTMLElement }>} apps
 */
export function unmountVueFlowIslands(apps) {
  if (!Array.isArray(apps)) return;

  apps.forEach(({ app, el }) => {
    try {
      app.unmount();
    } catch {
      // Ignore teardown race conditions during route transitions.
    }
    el?.removeAttribute('data-vueflow-mounted');
  });

  apps.length = 0;
}
