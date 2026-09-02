<script setup>
import { computed, ref, inject, watch, onBeforeUnmount } from 'vue';
import axios from 'axios';

const props = defineProps({
    show: { type: Boolean, required: true },
});

// imported prévient le parent qu'il doit recharger ses listes.
const emit = defineEmits(['close', 'imported']);

const apiUrl = inject('apiUrl');

const ICONS = {
    source: 'bi bi-plug-fill',
    item: 'bi bi-box-seam-fill',
    view: 'bi bi-eye-fill',
    secret: 'bi bi-key-fill',
};

const file = ref(null);
const fileInput = ref(null);
const preview = ref(null);
// decisions suit l'ordre de preview.entries, une ligne pour une ligne.
const decisions = ref([]);
const report = ref(null);
const step = ref('select');
const passphrase = ref('');
const loading = ref(false);
const applying = ref(false);
const error = ref('');

function resetAll() {
    file.value = null;
    if (fileInput.value) {
        fileInput.value.value = '';
    }
    preview.value = null;
    decisions.value = [];
    report.value = null;
    step.value = 'select';
    passphrase.value = '';
    error.value = '';
}

const OUTCOMES = {
    created: { icon: 'bi bi-check-circle-fill', variant: 'text-success', label: 'created' },
    updated: { icon: 'bi bi-check-circle-fill', variant: 'text-primary', label: 'updated' },
    skipped: { icon: 'bi bi-dash-circle', variant: 'text-secondary', label: 'skipped' },
};

// done : l'import est passé, la table ne se pilote plus, elle se lit.
const done = computed(() => Boolean(report.value));

// results indexe le rapport par nom d'archive, la clé des lignes du tableau.
const results = computed(() => {
    const map = {};
    for (const result of report.value?.results || []) {
        map[`${result.type}:${result.name}`] = result;
    }
    return map;
});

function outcomeFor(decision) {
    return results.value[`${decision.type}:${decision.name}`] || null;
}

const counts = computed(() => {
    const totals = { created: 0, updated: 0, skipped: 0 };
    for (const result of report.value?.results || []) {
        totals[result.outcome] = (totals[result.outcome] || 0) + 1;
    }
    return totals;
});

function messageFrom(requestError, fallback) {
    return requestError?.response?.data?.error || fallback;
}

async function onFileSelected(event) {
    const [selected] = event.target.files || [];
    preview.value = null;
    decisions.value = [];
    report.value = null;
    error.value = '';
    file.value = selected || null;
    if (!selected) {
        return;
    }

    loading.value = true;
    try {
        const body = new FormData();
        body.append('file', selected);
        const response = await axios.post(`${apiUrl}/import/preview`, body);
        preview.value = response.data;
        decisions.value = (response.data.entries || []).map((entry) => ({
            type: entry.type,
            name: entry.name,
            action: entry.action,
            as: entry.as,
        }));
    } catch (requestError) {
        console.error('Import preview failed', requestError);
        error.value = messageFrom(requestError, 'edit.bundle.errors.preview');
        file.value = null;
    } finally {
        loading.value = false;
    }
}

function entryFor(index) {
    return preview.value?.entries?.[index] || {};
}

// finalName est le nom sous lequel la ligne sera écrite : le nom de l'archive
// pour une mise à jour, le nom saisi pour une création.
function finalName(decision) {
    return decision.action === 'update' ? decision.name : decision.as;
}

const duplicates = computed(() => {
    const seen = {};
    const clashing = {};
    for (const decision of decisions.value) {
        if (decision.action === 'skip') {
            continue;
        }
        const entryKey = `${decision.type}:${finalName(decision)}`;
        if (seen[entryKey]) {
            clashing[entryKey] = true;
        }
        seen[entryKey] = true;
    }
    return clashing;
});

function isDuplicate(decision) {
    if (decision.action === 'skip') {
        return false;
    }
    return Boolean(duplicates.value[`${decision.type}:${finalName(decision)}`]);
}

function isNameMissing(decision) {
    return decision.action === 'create' && !String(decision.as || '').trim();
}

const invalidCount = computed(
    () => decisions.value.filter((decision) => isDuplicate(decision) || isNameMissing(decision)).length
);

const plannedCount = computed(
    () => decisions.value.filter((decision) => decision.action !== 'skip').length
);

const needsPassphrase = computed(
    () => Boolean(preview.value?.needs_passphrase) &&
        decisions.value.some((decision) => decision.type === 'secret' && decision.action !== 'skip')
);

const canApply = computed(
    () => Boolean(preview.value) && invalidCount.value === 0 && plannedCount.value > 0 && !applying.value
);

function setAllActions(action) {
    decisions.value = decisions.value.map((decision, index) => {
        const entry = entryFor(index);
        // Une entité sans homonyme local ne peut pas être une mise à jour.
        if (action === 'update' && !entry.collides) {
            return { ...decision, action: 'create', as: entry.as };
        }
        return { ...decision, action };
    });
}

function requestImport() {
    if (needsPassphrase.value) {
        step.value = 'passphrase';
        return;
    }
    runImport();
}

async function runImport() {
    applying.value = true;
    error.value = '';
    report.value = null;

    try {
        const body = new FormData();
        body.append('file', file.value);
        body.append('decisions', JSON.stringify(decisions.value));
        body.append('passphrase', passphrase.value);

        const response = await axios.post(`${apiUrl}/import/apply`, body);
        report.value = response.data;
        passphrase.value = '';
        step.value = 'select';
        emit('imported');
    } catch (requestError) {
        console.error('Import failed', requestError);
        error.value = messageFrom(requestError, 'edit.bundle.errors.import');
        step.value = 'select';
    } finally {
        applying.value = false;
    }
}

watch(
    () => props.show,
    (opened) => {
        document.body.classList.toggle('modal-open', opened);
        if (opened) {
            resetAll();
        }
    },
    { immediate: true }
);

onBeforeUnmount(() => {
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
                        <i class="bi bi-box-arrow-in-up"></i>
                        <span v-if="step === 'select'">{{ $t('edit.bundle.import.header') }}</span>
                        <span v-else>{{ $t('edit.bundle.import.passphrasetitle') }}</span>
                    </h5>
                    <button type="button" class="btn-close" aria-label="Close" @click="emit('close')"></button>
                </div>

                <div class="modal-body">
                    <div v-if="step === 'select'" class="d-flex flex-column gap-3">
                        <div>
                            <label for="bundleImportFile" class="form-label mb-1">
                                {{ $t('edit.bundle.import.file') }}
                            </label>
                            <input id="bundleImportFile" ref="fileInput" type="file" accept=".zip,application/zip"
                                class="form-control" @change="onFileSelected">
                        </div>

                        <div v-if="loading" class="d-flex align-items-center gap-2 text-secondary">
                            <span class="spinner-border spinner-border-sm" role="status"></span>
                            <span>{{ $t('edit.bundle.import.reading') }}</span>
                        </div>

                        <div v-if="error" class="alert alert-danger mb-0 py-2 px-3">
                            {{ error.startsWith('edit.') ? $t(error) : error }}
                        </div>

                        <div v-if="preview?.warnings?.length"
                            class="alert edit-bundle-alert alert-warning mb-0 py-2 px-3 small">
                            <p class="fw-semibold mb-1">{{ $t('edit.bundle.import.exportwarnings') }}</p>
                            <ul class="mb-0 ps-3">
                                <li v-for="warning in preview.warnings" :key="warning">{{ warning }}</li>
                            </ul>
                        </div>

                        <template v-if="preview">
                            <div v-if="!done" class="d-flex flex-wrap align-items-center gap-2">
                                <span class="small text-secondary">{{ $t('edit.bundle.import.setall') }}</span>
                                <div class="btn-group btn-group-sm">
                                    <button type="button" class="btn btn-outline-primary"
                                        @click="setAllActions('update')">
                                        {{ $t('edit.bundle.import.actions.update') }}
                                    </button>
                                    <button type="button" class="btn btn-outline-success"
                                        @click="setAllActions('create')">
                                        {{ $t('edit.bundle.import.actions.create') }}
                                    </button>
                                    <button type="button" class="btn btn-outline-secondary"
                                        @click="setAllActions('skip')">
                                        {{ $t('edit.bundle.import.actions.skip') }}
                                    </button>
                                </div>
                                <span class="ms-auto badge rounded-pill edit-bundle-chip text-bg-primary">
                                    {{ $t('edit.bundle.import.planned') }}: {{ plannedCount }} / {{ decisions.length }}
                                </span>
                            </div>

                            <div class="table-responsive edit-bundle-table-wrap">
                                <table class="table align-middle mb-0 edit-bundle-table">
                                    <thead>
                                        <tr>
                                            <th scope="col">{{ $t('edit.bundle.import.entity') }}</th>
                                            <th scope="col">{{ $t('edit.bundle.import.action') }}</th>
                                            <th scope="col">{{ $t('edit.bundle.import.finalname') }}</th>
                                            <th scope="col">{{ $t('edit.bundle.import.impact') }}</th>
                                            <th v-if="done" scope="col">
                                                {{ $t('edit.bundle.import.result') }}
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody class="table-group-divider">
                                        <tr v-for="(decision, index) in decisions"
                                            :key="`${decision.type}:${decision.name}`">
                                            <td>
                                                <div class="d-flex align-items-center gap-2">
                                                    <i :class="ICONS[decision.type]"></i>
                                                    <span class="text-truncate">{{ decision.name }}</span>
                                                </div>
                                                <span v-if="entryFor(index).collides"
                                                    class="badge rounded-pill text-bg-warning mt-1">
                                                    {{ $t('edit.bundle.import.collides') }}
                                                </span>
                                                <span v-else class="badge rounded-pill text-bg-success mt-1">
                                                    {{ $t('edit.bundle.import.new') }}
                                                </span>
                                            </td>
                                            <td>
                                                <select class="form-select form-select-sm" :disabled="done"
                                                    v-model="decision.action">
                                                    <option v-if="entryFor(index).collides" value="update">
                                                        {{ $t('edit.bundle.import.actions.update') }}
                                                    </option>
                                                    <option value="create">
                                                        {{ $t('edit.bundle.import.actions.create') }}
                                                    </option>
                                                    <option value="skip">
                                                        {{ $t('edit.bundle.import.actions.skip') }}
                                                    </option>
                                                </select>
                                            </td>
                                            <td>
                                                <input v-if="decision.action === 'create'" type="text"
                                                    class="form-control form-control-sm" :disabled="done"
                                                    :class="{ 'is-invalid': isDuplicate(decision) || isNameMissing(decision) }"
                                                    v-model="decision.as">
                                                <span v-else-if="decision.action === 'update'" class="text-secondary">
                                                    {{ decision.name }}
                                                </span>
                                                <span v-else class="text-secondary">-</span>
                                                <div v-if="isDuplicate(decision)" class="invalid-feedback d-block">
                                                    {{ $t('edit.bundle.import.duplicatename') }}
                                                </div>
                                                <div v-else-if="isNameMissing(decision)"
                                                    class="invalid-feedback d-block">
                                                    {{ $t('edit.bundle.import.emptyname') }}
                                                </div>
                                            </td>
                                            <td>
                                                <!-- Écraser une entité change le comportement de
                                                     tout ce qui en dépend localement. -->
                                                <span
                                                    v-if="decision.action === 'update' && entryFor(index).dependents?.length"
                                                    class="small text-warning-emphasis">
                                                    <i class="bi bi-exclamation-triangle me-1"></i>
                                                    {{ $t('edit.bundle.import.affects', {
                                                        count: entryFor(index).dependents.length
                                                    }) }}
                                                    <span class="d-block text-secondary">
                                                        {{ entryFor(index).dependents.join(', ') }}
                                                    </span>
                                                </span>
                                                <span v-else class="text-secondary">-</span>
                                            </td>
                                            <td v-if="done">
                                                <span v-if="outcomeFor(decision)"
                                                    class="d-inline-flex align-items-center gap-2"
                                                    :class="OUTCOMES[outcomeFor(decision).outcome].variant">
                                                    <i :class="OUTCOMES[outcomeFor(decision).outcome].icon"></i>
                                                    <span class="small">
                                                        {{ $t(`edit.bundle.import.outcomes.${outcomeFor(decision).outcome}`) }}
                                                    </span>
                                                </span>
                                                <span v-else class="text-secondary">-</span>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </template>

                        <div v-if="report" class="edit-bundle-panel p-3 d-flex flex-column gap-2">
                            <h6 class="mb-0">
                                <i class="bi bi-check-circle me-2"></i>{{ $t('edit.bundle.import.reporttitle') }}
                            </h6>
                            <div class="d-flex flex-wrap gap-2">
                                <span class="badge rounded-pill text-bg-success">
                                    {{ $t('edit.bundle.import.created') }}: {{ counts.created }}
                                </span>
                                <span class="badge rounded-pill text-bg-primary">
                                    {{ $t('edit.bundle.import.updated') }}: {{ counts.updated }}
                                </span>
                                <span class="badge rounded-pill text-bg-secondary">
                                    {{ $t('edit.bundle.import.skipped') }}: {{ counts.skipped }}
                                </span>
                            </div>
                            <ul v-if="report.warnings.length" class="mb-0 ps-3 small text-warning-emphasis">
                                <li v-for="warning in report.warnings" :key="warning">{{ warning }}</li>
                            </ul>
                        </div>
                    </div>

                    <div v-else class="d-flex flex-column gap-3">
                        <p class="mb-0 text-secondary small">{{ $t('edit.bundle.import.passphrasehelp') }}</p>
                        <div>
                            <label for="bundleImportPassphrase" class="form-label">
                                {{ $t('edit.bundle.passphrase') }}
                            </label>
                            <input id="bundleImportPassphrase" type="password" class="form-control"
                                autocomplete="current-password" v-model="passphrase">
                        </div>
                    </div>
                </div>

                <div class="modal-footer border-0 bg-body-tertiary">
                    <span v-if="done" class="me-auto small text-success">
                        <i class="bi bi-check-circle-fill me-1"></i>{{ $t('edit.bundle.import.reporttitle') }}
                    </span>
                    <span v-else-if="step === 'select' && invalidCount > 0" class="me-auto small text-danger">
                        {{ $t('edit.bundle.import.fixfirst', { count: invalidCount }) }}
                    </span>
                    <span v-else-if="step === 'select' && needsPassphrase" class="me-auto small text-secondary">
                        <i class="bi bi-shield-lock me-1"></i>{{ $t('edit.bundle.import.secretshint') }}
                    </span>

                    <template v-if="step === 'select'">
                        <button type="button" class="btn" :class="done ? 'btn-primary' : 'btn-secondary'"
                            @click="emit('close')">
                            {{ $t('global.close') }}
                        </button>
                        <!-- L'import est fait : plus rien à déclencher, la table
                             ne sert qu'à lire ce qui s'est passé. -->
                        <button v-if="!done" type="button" class="btn btn-primary d-inline-flex align-items-center gap-2"
                            :disabled="!canApply" @click="requestImport()">
                            <span v-if="applying" class="spinner-border spinner-border-sm" role="status"></span>
                            <i v-else class="bi bi-box-arrow-in-up"></i>
                            <span>{{ $t('edit.bundle.import.run') }}</span>
                        </button>
                    </template>

                    <template v-else>
                        <button type="button" class="btn btn-secondary" @click="step = 'select'">
                            {{ $t('edit.bundle.back') }}
                        </button>
                        <button type="button" class="btn btn-primary d-inline-flex align-items-center gap-2"
                            :disabled="passphrase.length === 0 || applying" @click="runImport()">
                            <span v-if="applying" class="spinner-border spinner-border-sm" role="status"></span>
                            <i v-else class="bi bi-box-arrow-in-up"></i>
                            <span>{{ $t('edit.bundle.import.run') }}</span>
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
