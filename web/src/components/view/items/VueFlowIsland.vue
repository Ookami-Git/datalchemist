<script setup>
// --- Imports Vue & Libs ---
import { computed, inject, onMounted, onBeforeUnmount, nextTick, ref } from 'vue';
import { VueFlow } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Controls } from '@vue-flow/controls';
import { MiniMap } from '@vue-flow/minimap';
import { resolveHeight } from '@/utils/vueFlowSize.js';

// CSS importé ici (et non dans main.js) pour rester dans le chunk asynchrone :
// les overrides dark vivent dans scss/_vueflow.scss et sont plus spécifiques,
// l'ordre de chargement ne les met donc pas en danger.
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import '@vue-flow/controls/dist/style.css';
import '@vue-flow/minimap/dist/style.css';

/**
 * Registre des composants de nœuds/arêtes personnalisés.
 *
 * Les `nodeTypes` de Vue Flow attendent des composants, ce qui n'est pas
 * sérialisable en JSON. Le template déclare donc un nom :
 *   "nodeTypes": { "custom": "statusCard" }
 * et ce nom est résolu ici. Ajouter une entrée pour exposer un nœud maison.
 */
const NODE_TYPE_REGISTRY = {};
const EDGE_TYPE_REGISTRY = {};

// Classe posée par viewGrid sur les widgets dont la taille suit le contenu.
const AUTO_RESIZE_CLASS = '.auto-resize-enabled';

// --- Props ---
const props = defineProps({
  config: { type: Object, required: true }, // sortie de parseFlowConfig()
  theme: { type: String, default: 'light' },
  hostElement: { type: Object, default: null }, // conteneur [data-vueflow] d'origine
});

// --- Injections (fournies par l'app hôte au montage de l'îlot) ---
const resizeWidget = inject('resizeWidget', null);

// --- Réactifs ---
const rootEl = ref(null);
const flowRef = ref(null);
const nodes = computed(() => props.config.nodes ?? []);
const edges = computed(() => props.config.edges ?? []);
const isDark = computed(() => props.theme === 'dark');

// Le mode peut être dégradé au montage si la configuration est contradictoire.
const sizing = ref({ ...props.config.sizing });
const height = ref(resolveHeight(props.config.sizing));

const containerStyle = computed(() => ({
  height: `${Math.round(height.value)}px`,
  width: '100%',
}));

/**
 * Résout un mapping { alias: "nomDansLeRegistre" } vers { alias: Composant }.
 * Les noms inconnus sont ignorés avec un warning plutôt que de casser le rendu.
 */
function resolveTypes(mapping, registry, label) {
  if (!mapping || typeof mapping !== 'object') return undefined;

  const resolved = {};
  Object.entries(mapping).forEach(([alias, name]) => {
    const component = typeof name === 'string' ? registry[name] : null;
    if (component) {
      resolved[alias] = component;
    } else {
      console.warn(`Vue Flow: unknown ${label} "${name}" for alias "${alias}"`);
    }
  });

  return Object.keys(resolved).length ? resolved : undefined;
}

// Props natives de <VueFlow> : tout ce que le template a fourni hors clés réservées,
// avec les types custom résolus en composants.
const flowProps = computed(() => {
  const { nodeTypes, edgeTypes, ...rest } = props.config.flowProps ?? {};

  const resolvedNodeTypes = resolveTypes(nodeTypes, NODE_TYPE_REGISTRY, 'node type');
  const resolvedEdgeTypes = resolveTypes(edgeTypes, EDGE_TYPE_REGISTRY, 'edge type');

  return {
    ...rest,
    ...(resolvedNodeTypes ? { nodeTypes: resolvedNodeTypes } : {}),
    ...(resolvedEdgeTypes ? { edgeTypes: resolvedEdgeTypes } : {}),
  };
});

// --- Plugins optionnels ---
const showBackground = computed(() => props.config.background !== false);
const backgroundProps = computed(() => {
  const raw = props.config.background;
  const base = raw && typeof raw === 'object' ? { ...raw } : {};
  if (!base.patternColor) base.patternColor = isDark.value ? '#495057' : '#b1b1b7';
  return base;
});

const showControls = computed(() => props.config.controls !== false);
const showMiniMap = computed(() => Boolean(props.config.minimap));
const miniMapProps = computed(() =>
  props.config.minimap && typeof props.config.minimap === 'object' ? props.config.minimap : {}
);

// --- Dimensionnement ---

/**
 * Hauteur restante dans le boîtier parent, pour le mode `fill`.
 *
 * On remet notre propre hauteur à zéro le temps de la mesure : en layout row la
 * hauteur de la carte dépend de son contenu, donc de nous. Sans cette mise à
 * zéro la mesure se contenterait de renvoyer notre hauteur courante et la
 * moindre variation serait figée définitivement.
 *
 * @returns {number|null} pixels disponibles, ou null si aucun boîtier trouvé
 */
function measureAvailableHeight() {
  const el = rootEl.value;
  if (!el) return null;

  const box = el.closest(props.config.fillTarget);
  if (!box) return null;

  const previousHeight = el.style.height;
  el.style.height = '0px';

  const boxRect = box.getBoundingClientRect();
  const elRect = el.getBoundingClientRect();
  const paddingBottom = Number.parseFloat(getComputedStyle(box).paddingBottom) || 0;

  el.style.height = previousHeight;

  return boxRect.bottom - paddingBottom - elRect.top;
}

let frame = null;
let appliedHeight = null;

function applyHeight() {
  const el = rootEl.value;
  if (!el) return;

  const available = sizing.value.mode === 'fill' ? measureAvailableHeight() : null;
  const next = resolveHeight(sizing.value, { width: el.clientWidth, available });

  // Seuil du pixel : coupe la boucle observer -> hauteur -> observer.
  if (appliedHeight !== null && Math.abs(next - appliedHeight) < 1) return;
  appliedHeight = next;
  height.value = next;

  nextTick(() => {
    // En mode fill, c'est le widget qui contraint le canvas : le pousser à se
    // redimensionner sur son contenu créerait une dépendance circulaire.
    if (resizeWidget && sizing.value.mode !== 'fill') resizeWidget();
    if (props.config.fitViewOnResize) {
      try {
        flowRef.value?.fitView();
      } catch {
        // Le canvas peut ne pas être encore prêt lors du premier passage.
      }
    }
  });
}

function scheduleApplyHeight() {
  if (frame !== null) return;
  frame = requestAnimationFrame(() => {
    frame = null;
    applyHeight();
  });
}

let observer = null;

// --- Cycle de vie ---
onMounted(async () => {
  await nextTick();

  // `fill` attend une hauteur imposée par le parent, `autoResize` fait dépendre
  // la hauteur du widget de son contenu : les deux ensemble ne convergent pas.
  if (sizing.value.mode === 'fill' && rootEl.value?.closest(AUTO_RESIZE_CLASS)) {
    console.warn(
      'Vue Flow: height "fill" cannot be used in a grid widget with autoResize enabled; ' +
      'falling back to a fixed height. Use aspectRatio instead, or disable autoResize.'
    );
    sizing.value = { ...sizing.value, mode: 'fixed' };
  }

  // Une hauteur fixe ne dépend d'aucune mesure : pas d'observer dans ce mode.
  if (sizing.value.mode !== 'fixed' && typeof ResizeObserver === 'function') {
    observer = new ResizeObserver(scheduleApplyHeight);
    // La largeur vient de notre propre boîte...
    if (rootEl.value) observer.observe(rootEl.value);
    // ...la hauteur disponible, du boîtier parent (widget grid ou carte de ligne).
    if (sizing.value.mode === 'fill') {
      const box = rootEl.value?.closest(props.config.fillTarget);
      if (box) observer.observe(box);
    }
  }

  applyHeight();
});

onBeforeUnmount(() => {
  observer?.disconnect();
  observer = null;
  if (frame !== null) {
    cancelAnimationFrame(frame);
    frame = null;
  }
});

// --- Pont d'événements vers le JS dynamique de l'item ---
// Les handlers Vue Flow ne sont pas exprimables en JSON : on les rediffuse en
// CustomEvent sur le conteneur, écoutables depuis le champ `javascript` de l'item.
function bridge(name, detail) {
  if (!props.hostElement) return;
  props.hostElement.dispatchEvent(
    new CustomEvent(`vueflow:${name}`, { detail, bubbles: true })
  );
}
</script>

<template>
  <div ref="rootEl" class="dc-vueflow" :class="{ 'dc-vueflow--dark': isDark }" :style="containerStyle">
    <VueFlow ref="flowRef" :nodes="nodes" :edges="edges" v-bind="flowProps"
      @node-click="bridge('nodeClick', { node: $event.node })"
      @node-double-click="bridge('nodeDoubleClick', { node: $event.node })"
      @edge-click="bridge('edgeClick', { edge: $event.edge })" @connect="bridge('connect', $event)"
      @pane-click="bridge('paneClick', {})">
      <Background v-if="showBackground" v-bind="backgroundProps" />
      <Controls v-if="showControls" />
      <MiniMap v-if="showMiniMap" v-bind="miniMapProps" />
    </VueFlow>
  </div>
</template>
