<script setup>

import { computed, ref, watch } from 'vue';

const props = defineProps({
    snapshot: {
        type: Object,
        default: null
    }
});

const expanded = ref(false);

const total = computed(() => props.snapshot?.total || 0);
const done = computed(() => props.snapshot?.done || 0);
const errors = computed(() => props.snapshot?.errors || 0);
const sources = computed(() => props.snapshot?.sources || []);
const percent = computed(() => Math.max(0, Math.min(100, props.snapshot?.percent || 0)));

// Rien à afficher tant que le backend n'a pas annoncé de source à charger.
const visible = computed(() => total.value > 0);

const RADIUS = 19;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;
const dashOffset = computed(() => CIRCUMFERENCE * (1 - percent.value / 100));

const running = computed(() => sources.value.filter((source) => source.status === 'running'));

const formatDuration = (milliseconds) => {
    if (!milliseconds) return '';
    if (milliseconds < 1000) return `${milliseconds} ms`;
    return `${(milliseconds / 1000).toFixed(milliseconds < 10000 ? 1 : 0)} s`;
};

const statusIcon = (status) => {
    switch (status) {
        case 'done': return 'bi-check-circle-fill';
        case 'error': return 'bi-exclamation-triangle-fill';
        case 'running': return 'bi-arrow-repeat';
        default: return 'bi-clock';
    }
};

watch(visible, (isVisible) => {
    if (!isVisible) expanded.value = false;
});
</script>

<template>
    <div v-if="visible" class="dc-progress" :class="{ 'is-expanded': expanded }">
        <transition name="dc-progress-panel">
            <div v-if="expanded" class="dc-progress__panel" role="dialog" :aria-label="$t('dataprogress.title')">
                <header class="dc-progress__panel-header">
                    <span class="dc-progress__panel-title">{{ $t('dataprogress.title') }}</span>
                    <span class="dc-progress__panel-count">{{ $t('dataprogress.summary', { done, total }) }}</span>
                </header>

                <ul class="dc-progress__list">
                    <li v-for="source in sources" :key="source.name" class="dc-progress__source"
                        :class="`is-${source.status}`">
                        <i class="bi dc-progress__source-icon" :class="statusIcon(source.status)"
                            aria-hidden="true"></i>

                        <div class="dc-progress__source-body">
                            <div class="dc-progress__source-head">
                                <span class="dc-progress__source-name" :title="source.name">{{ source.name }}</span>
                                <span v-if="source.loop" class="dc-progress__loop">
                                    <i class="bi bi-arrow-repeat" aria-hidden="true"></i>
                                    {{ $t('dataprogress.loop') }}
                                </span>
                            </div>

                            <div class="dc-progress__source-meta">
                                <span v-if="source.loop && source.looptotal">
                                    {{ source.percent }}% ({{ source.loopdone }}/{{ source.looptotal }})
                                </span>
                                <span v-else>{{ $t(`dataprogress.status.${source.status}`) }}</span>
                                <span v-if="source.duration" class="dc-progress__duration">
                                    {{ formatDuration(source.duration) }}
                                </span>
                            </div>

                            <div v-if="source.loop && source.looptotal" class="dc-progress__bar">
                                <span class="dc-progress__bar-fill" :style="{ width: `${source.percent}%` }"></span>
                            </div>

                            <div v-if="source.error" class="dc-progress__error" :title="source.error">
                                {{ source.error }}
                            </div>
                        </div>
                    </li>
                </ul>
            </div>
        </transition>

        <button type="button" class="dc-progress__bubble" :aria-expanded="expanded"
            :aria-label="expanded ? $t('dataprogress.hide') : $t('dataprogress.show')"
            :title="running.length ? `${done}/${total} — ${running.map((source) => source.name).join(', ')}` : $t('dataprogress.summary', { done, total })"
            @click="expanded = !expanded">
            <span class="dc-progress__halo" aria-hidden="true"></span>

            <svg class="dc-progress__ring" viewBox="0 0 44 44" aria-hidden="true">
                <circle class="dc-progress__ring-track" cx="22" cy="22" :r="RADIUS" />
                <circle class="dc-progress__ring-bar" :class="{ 'has-error': errors }" cx="22" cy="22" :r="RADIUS"
                    :stroke-dasharray="CIRCUMFERENCE" :stroke-dashoffset="dashOffset" />
            </svg>

            <span class="dc-progress__value">
                <span class="dc-progress__done">{{ done }}</span>
                <span class="dc-progress__separator">/</span>
                <span class="dc-progress__total">{{ total }}</span>
            </span>
        </button>
    </div>
</template>

<style lang="scss" scoped>
.dc-progress {
    --progress-primary: #2f7cf4;
    --progress-secondary: #17b897;
    --progress-danger: #dc3545;
    --progress-surface: rgba(255, 255, 255, 0.97);
    --progress-border: rgba(15, 30, 55, 0.12);
    --progress-track: rgba(15, 30, 55, 0.12);
    --progress-ink: #1f2b3d;
    --progress-muted: #667c99;
    --progress-glow: rgba(47, 124, 244, 0.32);
    --progress-shadow: rgba(15, 30, 55, 0.18);
    --progress-shadow-strong: rgba(15, 30, 55, 0.24);

    position: fixed;
    right: 1.5rem;
    bottom: 1.5rem;
    z-index: 1040;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
}

// Sans :global() : le compilateur scoped remplacerait tout le sélecteur par le
// contenu de :global(), les variables seraient posées sur <html> et la
// déclaration locale de .dc-progress les écraserait.
[data-bs-theme='dark'] .dc-progress {
    --progress-primary: #7ab7ff;
    --progress-secondary: #71e0bf;
    --progress-danger: #ff7b8a;
    --progress-surface: rgba(24, 31, 43, 0.97);
    --progress-border: rgba(255, 255, 255, 0.14);
    --progress-track: rgba(255, 255, 255, 0.16);
    --progress-ink: #eef4ff;
    --progress-muted: #a4b6cf;
    --progress-glow: rgba(122, 183, 255, 0.42);
    --progress-shadow: rgba(0, 0, 0, 0.45);
    --progress-shadow-strong: rgba(0, 0, 0, 0.55);
}

.dc-progress__bubble {
    position: relative;
    width: 66px;
    height: 66px;
    padding: 0;
    border: 1px solid var(--progress-border);
    border-radius: 999px;
    background: var(--progress-surface);
    box-shadow: 0 12px 28px var(--progress-shadow);
    color: var(--progress-ink);
    display: grid;
    place-items: center;
    cursor: pointer;
    animation: dc-progress-float 4.5s ease-in-out infinite;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.dc-progress__bubble:hover,
.dc-progress__bubble:focus-visible {
    transform: translateY(-2px) scale(1.04);
    box-shadow: 0 16px 34px var(--progress-shadow-strong);
}

.dc-progress__halo {
    position: absolute;
    inset: -9px;
    border-radius: 999px;
    background: radial-gradient(circle, var(--progress-glow) 0%, rgba(0, 0, 0, 0) 70%);
    animation: dc-progress-breathe 2.6s ease-in-out infinite;
    pointer-events: none;
}

.dc-progress__ring {
    position: absolute;
    inset: 4px;
    width: calc(100% - 8px);
    height: calc(100% - 8px);
    transform: rotate(-90deg);
}

.dc-progress__ring-track,
.dc-progress__ring-bar {
    fill: none;
    stroke-width: 3.5;
    stroke-linecap: round;
}

.dc-progress__ring-track {
    stroke: var(--progress-track);
}

.dc-progress__ring-bar {
    stroke: var(--progress-primary);
    transition: stroke-dashoffset 0.35s ease;
}

.dc-progress__ring-bar.has-error {
    stroke: var(--progress-danger);
}

.dc-progress__value {
    position: relative;
    display: inline-flex;
    align-items: baseline;
    gap: 0.05rem;
    font-size: 0.9rem;
    font-weight: 600;
    line-height: 1;
    letter-spacing: -0.02em;
}

.dc-progress__separator,
.dc-progress__total {
    color: var(--progress-muted);
    font-size: 0.75rem;
    font-weight: 500;
}

.dc-progress__panel {
    width: min(22rem, calc(100vw - 3rem));
    margin-bottom: 0.75rem;
    padding: 0.75rem;
    border: 1px solid var(--progress-border);
    border-radius: 0.9rem;
    background: var(--progress-surface);
    box-shadow: 0 18px 40px var(--progress-shadow-strong);
    color: var(--progress-ink);
}

.dc-progress__panel-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.5rem;
    padding-bottom: 0.5rem;
    margin-bottom: 0.35rem;
    border-bottom: 1px solid var(--progress-border);
}

.dc-progress__panel-title {
    font-size: 0.85rem;
    font-weight: 600;
}

.dc-progress__panel-count {
    font-size: 0.72rem;
    color: var(--progress-muted);
    white-space: nowrap;
}

.dc-progress__list {
    max-height: min(22rem, 50vh);
    margin: 0;
    padding: 0;
    overflow-y: auto;
    list-style: none;
}

.dc-progress__source {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.35rem 0.15rem;
}

.dc-progress__source+.dc-progress__source {
    border-top: 1px dashed var(--progress-border);
}

.dc-progress__source-icon {
    flex: 0 0 auto;
    margin-top: 0.1rem;
    font-size: 0.85rem;
    color: var(--progress-muted);
}

.dc-progress__source.is-done .dc-progress__source-icon {
    color: var(--progress-secondary);
}

.dc-progress__source.is-error .dc-progress__source-icon {
    color: var(--progress-danger);
}

.dc-progress__source.is-running .dc-progress__source-icon {
    color: var(--progress-primary);
    animation: dc-progress-spin 1.1s linear infinite;
}

.dc-progress__source-body {
    min-width: 0;
    flex: 1 1 auto;
}

.dc-progress__source-head {
    display: flex;
    align-items: center;
    gap: 0.35rem;
}

.dc-progress__source-name {
    min-width: 0;
    font-size: 0.8rem;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.dc-progress__loop {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    padding: 0 0.35rem;
    border-radius: 999px;
    background: rgba(47, 124, 244, 0.16);
    background: color-mix(in srgb, var(--progress-primary) 16%, transparent);
    color: var(--progress-primary);
    font-size: 0.65rem;
    line-height: 1.4;
}

.dc-progress__source-meta {
    display: flex;
    gap: 0.4rem;
    font-size: 0.7rem;
    color: var(--progress-muted);
    font-variant-numeric: tabular-nums;
}

.dc-progress__duration::before {
    content: '·';
    margin-right: 0.3rem;
}

.dc-progress__bar {
    height: 3px;
    margin-top: 0.25rem;
    border-radius: 999px;
    background: var(--progress-track);
    overflow: hidden;
}

.dc-progress__bar-fill {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: var(--progress-primary);
    transition: width 0.3s ease;
}

.dc-progress__error {
    margin-top: 0.15rem;
    font-size: 0.68rem;
    color: var(--progress-danger);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.dc-progress-panel-enter-active,
.dc-progress-panel-leave-active {
    transition: opacity 0.18s ease, transform 0.18s ease;
}

.dc-progress-panel-enter-from,
.dc-progress-panel-leave-to {
    opacity: 0;
    transform: translateY(0.5rem) scale(0.98);
}

@keyframes dc-progress-float {

    0%,
    100% {
        transform: translateY(0);
    }

    50% {
        transform: translateY(-6px);
    }
}

@keyframes dc-progress-breathe {

    0%,
    100% {
        opacity: 0.45;
        transform: scale(0.92);
    }

    50% {
        opacity: 0.9;
        transform: scale(1.06);
    }
}

@keyframes dc-progress-spin {
    from {
        transform: rotate(0deg);
    }

    to {
        transform: rotate(360deg);
    }
}

@media (max-width: 575.98px) {
    .dc-progress {
        right: 1rem;
        bottom: 1rem;
    }
}

@media (prefers-reduced-motion: reduce) {

    .dc-progress__bubble,
    .dc-progress__halo,
    .dc-progress__source.is-running .dc-progress__source-icon {
        animation: none;
    }
}
</style>
