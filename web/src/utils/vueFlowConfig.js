import { normalizeSizing } from './vueFlowSize.js';

/**
 * Parsing et normalisation de la configuration d'un îlot Vue Flow.
 *
 * Module volontairement pur (aucun import Vue/DOM) pour être testable avec
 * `node --test`, cf. vueFlowIslands.js pour le montage.
 */

// Clés consommées par l'îlot lui-même, le reste part en v-bind sur <VueFlow>.
export const RESERVED_KEYS = [
  'nodes',
  'edges',
  'background',
  'controls',
  'minimap',
  // dimensionnement, cf. vueFlowSize.js
  'height',
  'aspectRatio',
  'minHeight',
  'maxHeight',
  'fillTarget',
  'fitViewOnResize',
];

// Boîtier de référence pour `height: "fill"`. `.card-body` existe dans les deux
// layouts (viewGrid et viewRow), ce qui évite un cas par layout.
export const DEFAULT_FILL_TARGET = '.card-body';

/**
 * @param {string|object} raw – JSON brut (textContent du conteneur) ou objet déjà parsé
 * @param {{ fallbackId?: string }} [options]
 * @returns {{ config: object, error?: undefined } | { error: string, config?: undefined }}
 */
export function parseFlowConfig(raw, { fallbackId = 'dc-flow' } = {}) {
  let parsed = raw;

  if (typeof raw === 'string') {
    const trimmed = raw.trim();
    if (!trimmed) return { error: 'empty configuration' };
    try {
      parsed = JSON.parse(trimmed);
    } catch (err) {
      return { error: `invalid JSON (${err.message})` };
    }
  }

  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return { error: 'configuration must be a JSON object' };
  }

  const nodes = Array.isArray(parsed.nodes) ? parsed.nodes : [];
  const edges = Array.isArray(parsed.edges) ? parsed.edges : [];

  // Tout ce qui n'est pas réservé est considéré comme une prop de <VueFlow>,
  // ce qui évite de maintenir un mapping et suit l'évolution de la lib.
  const flowProps = {};
  Object.entries(parsed).forEach(([key, value]) => {
    if (!RESERVED_KEYS.includes(key)) flowProps[key] = value;
  });
  // Un id distinct par instance est obligatoire dès que plusieurs flows coexistent.
  if (!flowProps.id) flowProps.id = fallbackId;

  const sizing = normalizeSizing(parsed);

  return {
    config: {
      nodes,
      edges,
      flowProps,
      sizing,
      fillTarget: typeof parsed.fillTarget === 'string' && parsed.fillTarget.trim()
        ? parsed.fillTarget.trim()
        : DEFAULT_FILL_TARGET,
      // Recadrer après un redimensionnement n'a de sens que si on cadrait au départ.
      fitViewOnResize: typeof parsed.fitViewOnResize === 'boolean'
        ? parsed.fitViewOnResize
        : Boolean(flowProps.fitViewOnInit),
      background: parsed.background,
      controls: parsed.controls,
      minimap: parsed.minimap,
    },
  };
}
