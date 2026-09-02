/**
 * Dimensionnement des îlots Vue Flow.
 *
 * Vue Flow ne rend rien dans un conteneur de hauteur nulle, mais les deux
 * layouts de vue n'offrent pas la même contrainte :
 *
 * - grid : `.grid-stack-item-content` est en position absolue → hauteur définie,
 *   sauf si le widget est en `autoResize` (là c'est le contenu qui pilote) ;
 * - row  : la carte est étirée par la colonne la plus haute de la ligne, donc
 *   hauteur définie mais dérivée du contenu voisin.
 *
 * D'où trois modes :
 * - `fixed`  (défaut)                    → hauteur en pixels, marche partout ;
 * - `aspect` (`aspectRatio`)             → hauteur déduite de la largeur, qui est
 *   toujours définie dans les deux layouts : c'est le mode responsive ;
 * - `fill`   (`height: "fill"`)          → hauteur restante dans le boîtier parent.
 *
 * Module volontairement pur (aucun import Vue/DOM) pour être testable.
 */

export const DEFAULT_HEIGHT_PX = 400;

// Garde-fou pour les modes calculés : évite un canvas inutilisable si la mesure
// renvoie une valeur dégénérée (colonne repliée, carte non encore étirée...).
export const DEFAULT_MIN_HEIGHT_PX = 120;

export const FILL = 'fill';

/**
 * Longueur CSS exploitable en arithmétique : nombre, "420" ou "420px".
 * Les unités relatives (%, vh, em) sont refusées car elles ne peuvent pas être
 * bornées par min/max ici — utiliser `aspectRatio` ou `fill` à la place.
 *
 * @param {unknown} value
 * @returns {number|null} pixels, ou null si non exploitable
 */
export function parseLength(value) {
  if (typeof value === 'number') return Number.isFinite(value) && value >= 0 ? value : null;
  if (typeof value !== 'string') return null;

  const match = value.trim().match(/^(\d+(?:\.\d+)?)(px)?$/i);
  if (!match) return null;

  const parsed = Number.parseFloat(match[1]);
  return Number.isFinite(parsed) ? parsed : null;
}

/**
 * Ratio largeur/hauteur : nombre (1.78), "16/9" ou "16:9".
 *
 * @param {unknown} value
 * @returns {number|null}
 */
export function parseAspectRatio(value) {
  if (typeof value === 'number') return Number.isFinite(value) && value > 0 ? value : null;
  if (typeof value !== 'string') return null;

  const trimmed = value.trim();
  const ratio = trimmed.match(/^(\d+(?:\.\d+)?)\s*[/:]\s*(\d+(?:\.\d+)?)$/);
  if (ratio) {
    const width = Number.parseFloat(ratio[1]);
    const height = Number.parseFloat(ratio[2]);
    return height > 0 && width > 0 ? width / height : null;
  }

  const single = Number.parseFloat(trimmed);
  return Number.isFinite(single) && single > 0 ? single : null;
}

/**
 * Déduit le mode et les bornes depuis la config brute du template.
 *
 * @param {object} raw – configuration JSON de l'îlot
 * @returns {{ mode: 'fixed'|'aspect'|'fill', fixedHeight: number,
 *             aspectRatio: number|null, minHeight: number|null, maxHeight: number|null }}
 */
export function normalizeSizing(raw = {}) {
  const aspectRatio = parseAspectRatio(raw.aspectRatio);
  const isFill = typeof raw.height === 'string' && raw.height.trim().toLowerCase() === FILL;

  let mode = 'fixed';
  if (isFill) mode = 'fill';
  else if (aspectRatio) mode = 'aspect';

  // En mode calculé, `height` reste utile comme valeur de repli avant la
  // première mesure (et si celle-ci échoue).
  const fixedHeight = parseLength(raw.height) ?? DEFAULT_HEIGHT_PX;

  const minHeight = parseLength(raw.minHeight)
    ?? (mode === 'fixed' ? null : DEFAULT_MIN_HEIGHT_PX);

  return {
    mode,
    fixedHeight,
    aspectRatio,
    minHeight,
    maxHeight: parseLength(raw.maxHeight),
  };
}

/**
 * Applique les bornes min/max.
 *
 * @param {number} height
 * @param {{ minHeight: number|null, maxHeight: number|null }} bounds
 * @returns {number}
 */
export function clampHeight(height, { minHeight = null, maxHeight = null } = {}) {
  let result = height;
  // max d'abord : une borne min explicite doit rester gagnante sur un max plus petit.
  if (maxHeight !== null) result = Math.min(result, maxHeight);
  if (minHeight !== null) result = Math.max(result, minHeight);
  return result;
}

/**
 * Hauteur finale de l'îlot, en pixels.
 *
 * @param {ReturnType<typeof normalizeSizing>} sizing
 * @param {{ width?: number, available?: number|null }} [measures]
 * @returns {number}
 */
export function resolveHeight(sizing, { width = 0, available = null } = {}) {
  let height = sizing.fixedHeight;

  if (sizing.mode === 'aspect' && sizing.aspectRatio && width > 0) {
    height = width / sizing.aspectRatio;
  } else if (sizing.mode === 'fill' && available !== null && available > 0) {
    height = available;
  }

  return clampHeight(height, sizing);
}
