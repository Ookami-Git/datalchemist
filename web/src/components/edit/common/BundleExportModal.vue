<script setup>
import { computed, ref, inject, watch, onBeforeUnmount } from 'vue';
import axios from 'axios';

const props = defineProps({
    show: { type: Boolean, required: true },
    // seed pré-sélectionne une entité : c'est l'export « solo » depuis une carte.
    seed: { type: Object, default: null },
});

const emit = defineEmits(['close']);

const apiUrl = inject('apiUrl');

const TYPES = [
    { key: 'source', endpoint: 'sources', icon: 'bi bi-plug-fill' },
    { key: 'item', endpoint: 'items', icon: 'bi bi-box-seam-fill' },
    { key: 'view', endpoint: 'views', icon: 'bi bi-eye-fill' },
    { key: 'secret', endpoint: 'secrets', icon: 'bi bi-key-fill' },
];

const BUCKETS = { source: 'sources', item: 'items', view: 'views', secret: 'secrets' };

const catalog = ref({ source: [], item: [], view: [], secret: [] });
// explicit : ce que l'utilisateur a coché lui-même.
const explicit = ref({ source: [], item: [], view: [], secret: [] });
// excluded : ce qu'il a décoché parmi les dépendances proposées. Sans cette
// mémoire, le calcul de fermeture suivant les recocherait aussitôt.
const excluded = ref({});
const resolution = ref(null);
const resolving = ref(false);
const exporting = ref(false);
const error = ref('');
const outcome = ref(null);
const step = ref('select');
const passphrase = ref('');
const passphraseConfirm = ref('');

let resolveTimer = null;

const key = (type, name) => `${type}:${name}`;

const pulled = computed(() => {
    const map = {};
    for (const requirement of resolution.value?.requirements || []) {
        if (!requirement.selected) {
            map[key(requirement.type, requirement.name)] = requirement;
        }
    }
    return map;
});

const isChecked = (type, name) => {
    if (excluded.value[key(type, name)]) {
        return false;
    }
    if (explicit.value[type].includes(name)) {
        return true;
    }
    return Boolean(pulled.value[key(type, name)]);
};

function toggle(type, name) {
    const entryKey = key(type, name);
    if (isChecked(type, name)) {
        explicit.value[type] = explicit.value[type].filter((entry) => entry !== name);
        if (pulled.value[entryKey]) {
            excluded.value = { ...excluded.value, [entryKey]: true };
        }
        return;
    }

    const { [entryKey]: removed, ...rest } = excluded.value;
    excluded.value = rest;
    if (!explicit.value[type].includes(name)) {
        explicit.value[type] = [...explicit.value[type], name];
    }
}

function selectAll(type) {
    explicit.value[type] = catalog.value[type].map((entity) => entity.name);
    excluded.value = Object.fromEntries(
        Object.entries(excluded.value).filter(([entryKey]) => !entryKey.startsWith(`${type}:`))
    );
}

function clearType(type) {
    explicit.value[type] = [];
    const additions = {};
    for (const entity of catalog.value[type]) {
        if (pulled.value[key(type, entity.name)]) {
            additions[key(type, entity.name)] = true;
        }
    }
    excluded.value = { ...excluded.value, ...additions };
}

function selectEverything() {
    for (const type of TYPES) {
        selectAll(type.key);
    }
}

function clearEverything() {
    explicit.value = { source: [], item: [], view: [], secret: [] };
    excluded.value = {};
    resolution.value = null;
}

// finalSelection est ce qui partira réellement : la fermeture calculée par le
// serveur, moins ce que l'utilisateur a explicitement retiré.
const finalSelection = computed(() => {
    const selection = { sources: [], items: [], views: [], secrets: [] };
    for (const requirement of resolution.value?.requirements || []) {
        if (excluded.value[key(requirement.type, requirement.name)]) {
            continue;
        }
        selection[BUCKETS[requirement.type]].push(requirement.name);
    }
    return selection;
});

const totalSelected = computed(() =>
    Object.values(finalSelection.value).reduce((total, names) => total + names.length, 0)
);

const includesSecrets = computed(() => finalSelection.value.secrets.length > 0);

const passphraseValid = computed(
    () => passphrase.value.length > 0 && passphrase.value === passphraseConfirm.value
);

const pulledCount = computed(() => Object.keys(pulled.value).length);

// Les secrets sont repérés en lisant les templates, pas via une relation
// déclarée : ils sont proposés, jamais imposés.
const uncertainNames = computed(() =>
    Object.values(pulled.value)
        .filter((requirement) => !requirement.certain)
        .map((requirement) => requirement.name)
);

function reasonFor(type, name) {
    return pulled.value[key(type, name)]?.pulled_by || [];
}

async function fetchCatalog() {
    try {
        const responses = await Promise.all(
            TYPES.map((type) => axios.get(`${apiUrl}/${type.endpoint}`))
        );
        const loaded = {};
        TYPES.forEach((type, index) => {
            loaded[type.key] = (responses[index].data || []).filter((entity) => entity && entity.name);
        });
        catalog.value = loaded;
    } catch (requestError) {
        console.error('Failed to fetch the export catalog', requestError);
        error.value = 'edit.bundle.errors.catalog';
    }
}

async function resolveSelection() {
    const anythingPicked = Object.values(explicit.value).some((names) => names.length > 0);
    if (!anythingPicked) {
        resolution.value = null;
        return;
    }

    resolving.value = true;
    try {
        const response = await axios.post(`${apiUrl}/export/resolve`, {
            sources: explicit.value.source,
            items: explicit.value.item,
            views: explicit.value.view,
            secrets: explicit.value.secret,
        });
        resolution.value = response.data;
        error.value = '';
    } catch (requestError) {
        console.error('Failed to resolve the selection', requestError);
        error.value = 'edit.bundle.errors.resolve';
    } finally {
        resolving.value = false;
    }
}

watch(
    explicit,
    () => {
        clearTimeout(resolveTimer);
        resolveTimer = setTimeout(resolveSelection, 300);
    },
    { deep: true }
);

function filenameFrom(response) {
    const disposition = response.headers['content-disposition'] || '';
    const match = disposition.match(/filename="([^"]+)"/);
    return match ? match[1] : 'datalchemist.zip';
}

// Le corps d'une réponse d'erreur est un Blob quand on demande un Blob : il faut
// le relire pour retrouver le message.
async function blobErrorMessage(requestError) {
    const data = requestError?.response?.data;
    if (!data || typeof data.text !== 'function') {
        return '';
    }
    try {
        return JSON.parse(await data.text()).error || '';
    } catch {
        return '';
    }
}

// Le passage par la passphrase est une étape à part entière de la modale plutôt
// qu'un champ glissé dans le formulaire : on ne protège pas une archive
// distraitement.
function requestExport() {
    if (includesSecrets.value) {
        step.value = 'passphrase';
        return;
    }
    runExport();
}

async function runExport() {
    exporting.value = true;
    error.value = '';
    outcome.value = null;

    try {
        const response = await axios.post(
            `${apiUrl}/export`,
            { selection: finalSelection.value, passphrase: passphrase.value },
            { responseType: 'blob' }
        );

        const url = URL.createObjectURL(response.data);
        const link = document.createElement('a');
        link.href = url;
        link.download = filenameFrom(response);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);

        outcome.value = {
            filename: link.download,
            warnings: Number(response.headers['x-export-warnings'] || 0),
        };
        step.value = 'select';
        passphrase.value = '';
        passphraseConfirm.value = '';
    } catch (requestError) {
        console.error('Export failed', requestError);
        error.value = (await blobErrorMessage(requestError)) || 'edit.bundle.errors.export';
        step.value = 'select';
    } finally {
        exporting.value = false;
    }
}

// L'ouverture repart d'un état propre : une sélection oubliée d'une session
// précédente serait exportée à l'insu de l'utilisateur.
watch(
    () => props.show,
    (opened) => {
        document.body.classList.toggle('modal-open', opened);
        if (!opened) {
            return;
        }

        clearEverything();
        error.value = '';
        outcome.value = null;
        step.value = 'select';
        passphrase.value = '';
        passphraseConfirm.value = '';

        fetchCatalog();
        if (props.seed?.type && props.seed?.name) {
            explicit.value[props.seed.type] = [props.seed.name];
        }
    },
    { immediate: true }
);

onBeforeUnmount(() => {
    clearTimeout(resolveTimer);
    document.body.classList.remove('modal-open');
});
</script>

<template>
    <div v-if="show" class="modal fade show d-block edit-bundle-modal" tabindex="-1" role="dialog" aria-modal="true"
        @click.self="emit('close')">
        <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable modal-xl" role="document">
            <div class="modal-content shadow-lg border-0">
                <div class="modal-header border-0 bg-body-tertiary">
                    <h5 class="modal-title d-flex align-items-center gap-2">
                        <i class="bi bi-box-arrow-down"></i>
                        <span v-if="step === 'select'">{{ $t('edit.bundle.export.header') }}</span>
                        <span v-else>{{ $t('edit.bundle.export.passphrasetitle') }}</span>
                    </h5>
                    <button type="button" class="btn-close" aria-label="Close" @click="emit('close')"></button>
                </div>

                <div class="modal-body">
                    <!-- Étape 1 : la sélection -->
                    <div v-if="step === 'select'" class="d-flex flex-column gap-3">
                        <div class="d-flex flex-wrap align-items-center gap-2">
                            <button type="button" class="btn btn-sm btn-outline-primary d-inline-flex gap-2"
                                @click="selectEverything()">
                                <i class="bi bi-check2-all"></i>
                                <span>{{ $t('edit.bundle.export.selectall') }}</span>
                            </button>
                            <button type="button" class="btn btn-sm btn-outline-secondary d-inline-flex gap-2"
                                :disabled="totalSelected === 0" @click="clearEverything()">
                                <i class="bi bi-x-lg"></i>
                                <span>{{ $t('edit.bundle.export.clear') }}</span>
                            </button>

                            <span class="ms-auto d-inline-flex align-items-center gap-2">
                                <span v-if="resolving" class="spinner-border spinner-border-sm text-secondary"
                                    role="status"></span>
                                <span class="badge rounded-pill edit-bundle-chip text-bg-primary">
                                    {{ $t('edit.bundle.export.selected') }}: {{ totalSelected }}
                                </span>
                                <span v-if="pulledCount > 0" class="badge rounded-pill edit-bundle-chip text-bg-info">
                                    <i class="bi bi-diagram-3 me-1"></i>{{ $t('edit.bundle.export.pulled') }}:
                                    {{ pulledCount }}
                                </span>
                            </span>
                        </div>

                        <div v-if="uncertainNames.length > 0"
                            class="alert edit-bundle-alert alert-warning mb-0 py-2 px-3 small">
                            <i class="bi bi-info-circle me-1"></i>
                            {{ $t('edit.bundle.export.uncertain') }} : {{ uncertainNames.join(', ') }}
                        </div>

                        <div v-if="resolution?.warnings?.length"
                            class="alert edit-bundle-alert alert-warning mb-0 py-2 px-3 small">
                            <ul class="mb-0 ps-3">
                                <li v-for="warning in resolution.warnings" :key="warning">{{ warning }}</li>
                            </ul>
                        </div>

                        <div class="row g-3">
                            <div v-for="type in TYPES" :key="type.key" class="col-12 col-lg-6">
                                <article class="edit-bundle-panel h-100 d-flex flex-column">
                                    <div class="edit-bundle-panel-head px-3 py-2 d-flex align-items-center gap-2">
                                        <i :class="type.icon"></i>
                                        <span class="fw-semibold">{{ $t(`edit.bundle.types.${type.key}`) }}</span>
                                        <span class="badge rounded-pill text-bg-secondary">
                                            {{ catalog[type.key].length }}
                                        </span>
                                        <div class="ms-auto btn-group btn-group-sm">
                                            <button type="button" class="btn btn-outline-primary"
                                                :title="$t('edit.bundle.export.selectall')"
                                                @click="selectAll(type.key)">
                                                <i class="bi bi-check2-all"></i>
                                            </button>
                                            <button type="button" class="btn btn-outline-secondary"
                                                :title="$t('edit.bundle.export.clear')" @click="clearType(type.key)">
                                                <i class="bi bi-x-lg"></i>
                                            </button>
                                        </div>
                                    </div>

                                    <div class="edit-bundle-list">
                                        <p v-if="catalog[type.key].length === 0"
                                            class="text-secondary small text-center py-3 mb-0">-</p>
                                        <label v-for="entity in catalog[type.key]" :key="entity.id ?? entity.name"
                                            class="edit-bundle-row d-flex align-items-center gap-2 px-3 py-2">
                                            <input class="form-check-input mt-0 flex-shrink-0" type="checkbox"
                                                :checked="isChecked(type.key, entity.name)"
                                                @change="toggle(type.key, entity.name)">
                                            <span class="text-truncate">{{ entity.name }}</span>
                                            <span v-if="reasonFor(type.key, entity.name).length"
                                                class="badge rounded-pill text-bg-info ms-auto flex-shrink-0"
                                                :title="reasonFor(type.key, entity.name).join(', ')">
                                                <i class="bi bi-diagram-3 me-1"></i>
                                                {{ $t('edit.bundle.export.pulledby') }}
                                            </span>
                                        </label>
                                    </div>
                                </article>
                            </div>
                        </div>

                        <div v-if="error" class="alert alert-danger mb-0 py-2 px-3">
                            {{ error.startsWith('edit.') ? $t(error) : error }}
                        </div>

                        <div v-if="outcome" class="alert alert-success mb-0 py-2 px-3">
                            <i class="bi bi-check-circle me-1"></i>
                            {{ $t('edit.bundle.export.done', { file: outcome.filename }) }}
                            <span v-if="outcome.warnings > 0">
                                — {{ $t('edit.bundle.export.donewarnings', { count: outcome.warnings }) }}
                            </span>
                        </div>
                    </div>

                    <!-- Étape 2 : la passphrase qui protégera les secrets -->
                    <div v-else class="d-flex flex-column gap-3">
                        <p class="mb-0 text-secondary small">{{ $t('edit.bundle.export.passphrasehelp') }}</p>

                        <div>
                            <label for="bundleExportPassphrase" class="form-label">
                                {{ $t('edit.bundle.passphrase') }}
                            </label>
                            <input id="bundleExportPassphrase" type="password" class="form-control"
                                autocomplete="new-password" v-model="passphrase">
                        </div>

                        <div>
                            <label for="bundleExportPassphraseConfirm" class="form-label">
                                {{ $t('edit.bundle.passphraseconfirm') }}
                            </label>
                            <input id="bundleExportPassphraseConfirm" type="password" class="form-control"
                                autocomplete="new-password"
                                :class="{ 'is-invalid': passphraseConfirm.length > 0 && !passphraseValid }"
                                v-model="passphraseConfirm">
                            <div class="invalid-feedback">{{ $t('edit.bundle.passphrasemismatch') }}</div>
                        </div>

                        <div class="alert alert-warning mb-0 py-2 px-3 small">
                            <i class="bi bi-exclamation-triangle me-1"></i>
                            {{ $t('edit.bundle.export.passphrasewarning') }}
                        </div>
                    </div>
                </div>

                <div class="modal-footer border-0 bg-body-tertiary">
                    <span v-if="step === 'select' && includesSecrets" class="me-auto small text-secondary">
                        <i class="bi bi-shield-lock me-1"></i>{{ $t('edit.bundle.export.secretshint') }}
                    </span>

                    <template v-if="step === 'select'">
                        <button type="button" class="btn btn-secondary" @click="emit('close')">
                            {{ $t('global.close') }}
                        </button>
                        <button type="button" class="btn btn-success d-inline-flex align-items-center gap-2"
                            :disabled="totalSelected === 0 || exporting" @click="requestExport()">
                            <span v-if="exporting" class="spinner-border spinner-border-sm" role="status"></span>
                            <i v-else class="bi bi-box-arrow-down"></i>
                            <span>{{ $t('edit.bundle.export.run') }}</span>
                        </button>
                    </template>

                    <template v-else>
                        <button type="button" class="btn btn-secondary" @click="step = 'select'">
                            {{ $t('edit.bundle.back') }}
                        </button>
                        <button type="button" class="btn btn-success d-inline-flex align-items-center gap-2"
                            :disabled="!passphraseValid || exporting" @click="runExport()">
                            <span v-if="exporting" class="spinner-border spinner-border-sm" role="status"></span>
                            <i v-else class="bi bi-box-arrow-down"></i>
                            <span>{{ $t('edit.bundle.export.run') }}</span>
                        </button>
                    </template>
                </div>
            </div>
        </div>
    </div>
    <div v-if="show" class="modal-backdrop fade show" style="z-index: 1040;"></div>
</template>

<style scoped>
.edit-bundle-modal {
    z-index: 1050;
    background: rgba(0, 0, 0, 0.15);
}
</style>
