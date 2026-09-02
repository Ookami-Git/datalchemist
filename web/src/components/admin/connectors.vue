<script setup>
import { computed, inject, onBeforeUnmount, onMounted, ref } from 'vue';
import axios from 'axios';

const apiUrl = inject('apiUrl');
const save = inject('save');

save.value.safe();

const PROVIDERS = [
    { id: 'gitlab', label: 'GitLab', icon: 'bi bi-gitlab' },
    { id: 'github', label: 'GitHub', icon: 'bi bi-github' },
    { id: 'other', label: 'Autre', icon: 'bi bi-git' },
];

const KIND_ICONS = {
    source: 'bi bi-plug-fill text-primary',
    item: 'bi bi-box-seam-fill text-success',
    view: 'bi bi-grid-1x2-fill text-info',
    secret: 'bi bi-shield-lock-fill text-warning',
};

const emptyForm = () => ({
    url: '',
    branch: 'main',
    directory: '',
    provider: 'gitlab',
    username: '',
    author_name: '',
    author_email: '',
    poll_interval: 60,
});

const emptyCredentials = () => ({ token: '', passphrase: '', webhook_secret: '' });
const emptyClear = () => ({ token: false, passphrase: false, webhook_secret: false });

const settings = ref(null);
const status = ref(null);
const form = ref(emptyForm());
const original = ref(null);
const credentials = ref(emptyCredentials());
const clear = ref(emptyClear());

const loading = ref(true);
const saving = ref(false);
const testing = ref(false);
const syncing = ref(false);
const enabling = ref(false);
const disabling = ref(false);
const error = ref('');
const notice = ref('');
const testResult = ref(null);

const enableModal = ref(null);
const disableModal = ref(false);
const conflictModal = ref(null);
const resolving = ref('');

let statusTimer = null;

const available = computed(() => status.value?.available !== false);
const enabled = computed(() => Boolean(status.value?.enabled));
const conflicts = computed(() => status.value?.conflicts || []);
const warnings = computed(() => status.value?.warnings || []);

const credentialsDirty = computed(() =>
    Object.values(credentials.value).some((value) => value !== '') ||
    Object.values(clear.value).some(Boolean)
);

const formDirty = computed(() => {
    if (!original.value) {
        return false;
    }
    return Object.keys(form.value).some((key) => `${form.value[key]}` !== `${original.value[key]}`);
});

const dirty = computed(() => formDirty.value || credentialsDirty.value);

const stateChip = computed(() => {
    if (!status.value) {
        return { key: 'disabled', classes: 'text-bg-secondary', icon: 'bi bi-hourglass' };
    }
    if (!enabled.value) {
        return { key: 'disabled', classes: 'text-bg-secondary', icon: 'bi bi-pause-circle' };
    }
    if (status.value.running) {
        return { key: 'running', classes: 'text-bg-info', icon: 'bi bi-arrow-repeat' };
    }
    if (status.value.last_error) {
        return { key: 'error', classes: 'text-bg-danger', icon: 'bi bi-x-octagon-fill' };
    }
    if (conflicts.value.length > 0) {
        return { key: 'conflicts', classes: 'text-bg-warning', icon: 'bi bi-exclamation-triangle-fill' };
    }
    return { key: 'enabled', classes: 'text-bg-success', icon: 'bi bi-check-circle-fill' };
});

const webhookUrl = computed(() => `${window.location.origin}${window.location.pathname}api/webhook/git`);

const providerIcon = computed(() =>
    (PROVIDERS.find((provider) => provider.id === form.value.provider) || PROVIDERS[2]).icon
);

function clone(data) {
    return JSON.parse(JSON.stringify(data));
}

function messageFrom(requestError, fallback) {
    return requestError?.response?.data?.error || requestError?.message || fallback;
}

function applySettings(data) {
    settings.value = data;
    form.value = { ...emptyForm(), ...(data?.config || {}) };
    original.value = clone(form.value);
    credentials.value = emptyCredentials();
    clear.value = emptyClear();
}

function credentialsPatch() {
    const patch = {};
    for (const field of Object.keys(credentials.value)) {
        if (credentials.value[field] !== '') {
            patch[field] = credentials.value[field];
        } else if (clear.value[field]) {
            patch[field] = '';
        }
    }
    return patch;
}

function requestBody() {
    return {
        config: { ...form.value, poll_interval: Number(form.value.poll_interval) || 60 },
        credentials: credentialsPatch(),
    };
}

async function fetchAll() {
    loading.value = true;
    try {
        const response = await axios.get(`${apiUrl}/connector/git`);
        applySettings(response.data.settings);
        status.value = response.data.status;
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Impossible de charger le connecteur');
    } finally {
        loading.value = false;
    }
}

async function fetchStatus() {
    try {
        const response = await axios.get(`${apiUrl}/connector/git/status`);
        status.value = response.data;
    } catch (requestError) {
        console.error('Failed to refresh git connector status', requestError);
    }
}

async function saveSettings() {
    saving.value = true;
    error.value = '';
    notice.value = '';
    try {
        const response = await axios.put(`${apiUrl}/connector/git`, requestBody());
        applySettings(response.data.settings);
        status.value = response.data.status;
        notice.value = 'saved';
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Enregistrement impossible');
    } finally {
        saving.value = false;
    }
}

async function testConnection() {
    testing.value = true;
    testResult.value = null;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/test`, requestBody());
        testResult.value = response.data;
    } catch (requestError) {
        testResult.value = { reachable: false, error: messageFrom(requestError, 'Test impossible') };
    } finally {
        testing.value = false;
    }
}

async function requestEnable() {
    enabling.value = true;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/test`, { config: form.value, credentials: {} });
        const probe = response.data;
        if (!probe.reachable) {
            error.value = probe.error || 'Dépôt inaccessible';
            return;
        }
        if (probe.empty) {
            await enable('');
            return;
        }
        enableModal.value = { probe, direction: 'merge' };
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Activation impossible');
    } finally {
        enabling.value = false;
    }
}

async function enable(direction) {
    enabling.value = true;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/enable`, { direction });
        status.value = response.data;
        enableModal.value = null;
        await fetchAll();
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Activation impossible');
        await fetchStatus();
    } finally {
        enabling.value = false;
    }
}

function confirmEnable() {
    const direction = enableModal.value?.direction === 'merge' ? '' : enableModal.value?.direction;
    enable(direction);
}

async function disable() {
    disabling.value = true;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/disable`);
        status.value = response.data;
        disableModal.value = false;
        await fetchAll();
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Désactivation impossible');
    } finally {
        disabling.value = false;
    }
}

async function syncNow() {
    syncing.value = true;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/sync`);
        status.value = response.data;
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Synchronisation impossible');
        await fetchStatus();
    } finally {
        syncing.value = false;
    }
}

async function openConflict(conflict) {
    try {
        const response = await axios.get(`${apiUrl}/connector/git/conflict/${conflict.kind}/${conflict.id}`);
        conflictModal.value = response.data;
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Détail du conflit indisponible');
    }
}

async function resolve(conflict, keep) {
    resolving.value = `${conflict.kind}:${conflict.id}:${keep}`;
    error.value = '';
    try {
        const response = await axios.post(`${apiUrl}/connector/git/conflict/${conflict.kind}/${conflict.id}/resolve`, { keep });
        status.value = response.data;
        conflictModal.value = null;
    } catch (requestError) {
        error.value = messageFrom(requestError, 'Arbitrage impossible');
        await fetchStatus();
    } finally {
        resolving.value = '';
    }
}

function isResolving(conflict, keep) {
    return resolving.value === `${conflict.kind}:${conflict.id}:${keep}`;
}

function fileNames(detail) {
    const names = new Set([...Object.keys(detail?.local || {}), ...Object.keys(detail?.remote || {})]);
    return [...names].sort();
}

function editRoute(conflict) {
    switch (conflict.kind) {
        case 'source':
            return { name: 'editsource', params: { sourceid: conflict.id } };
        case 'item':
            return { name: 'edititem', params: { itemid: conflict.id } };
        case 'view':
            return { name: 'editview', params: { viewid: conflict.id } };
        default:
            return { name: 'edit', query: { tab: 'secrets' } };
    }
}

function formatDate(value) {
    if (!value) {
        return null;
    }
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

function shortCommit(hash) {
    return hash ? hash.slice(0, 10) : '';
}

function kindIcon(kind) {
    return KIND_ICONS[kind] || 'bi bi-question-circle';
}

async function copyWebhookUrl() {
    try {
        await navigator.clipboard.writeText(webhookUrl.value);
        notice.value = 'copied';
    } catch (copyError) {
        console.error('Clipboard unavailable', copyError);
    }
}

onMounted(async () => {
    await fetchAll();
    statusTimer = window.setInterval(fetchStatus, 5000);
});

onBeforeUnmount(() => {
    if (statusTimer) {
        window.clearInterval(statusTimer);
    }
});
</script>

<template>
    <section class="admin-connectors-page container-fluid px-0 py-1 py-lg-2">
        <div class="d-flex flex-column gap-3 gap-xxl-4">
            <header class="card admin-connectors-hero shadow-sm">
                <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
                    <div class="admin-connectors-hero-icon">
                        <i class="bi bi-plug-fill"></i>
                    </div>
                    <div class="flex-grow-1">
                        <p class="admin-connectors-kicker mb-1">{{ $t('admin.header') }}</p>
                        <h4 class="mb-1">{{ $t('admin.connectors.header') }}</h4>
                        <p class="mb-0 text-secondary">{{ $t('admin.connectors.subtitle') }}</p>
                    </div>
                    <span class="badge rounded-pill admin-connectors-state-chip" :class="stateChip.classes">
                        <i :class="stateChip.icon" class="me-1"></i>
                        {{ $t(`admin.connectors.git.state.${stateChip.key}`, { count: conflicts.length }) }}
                    </span>
                </div>
            </header>

            <div v-if="!available" class="alert alert-warning d-flex align-items-center gap-2 mb-0">
                <i class="bi bi-key-fill"></i>
                <span>{{ $t('admin.connectors.unavailable') }}</span>
            </div>

            <div v-if="error" class="alert alert-danger d-flex align-items-start gap-2 mb-0">
                <i class="bi bi-x-octagon-fill mt-1"></i>
                <span class="flex-grow-1">{{ error }}</span>
                <button type="button" class="btn-close" aria-label="Close" @click="error = ''"></button>
            </div>

            <div class="row g-3 g-xxl-4">
                <!-- Configuration -->
                <div class="col-12 col-xl-7">
                    <article class="card admin-connectors-panel shadow-sm h-100">
                        <div class="card-body p-4 d-flex flex-column gap-3">
                            <div class="d-flex align-items-center gap-2">
                                <i :class="providerIcon" class="fs-4"></i>
                                <h5 class="admin-connectors-panel-title mb-0">{{ $t('admin.connectors.git.title') }}</h5>
                            </div>
                            <p class="text-secondary small mb-0">{{ $t('admin.connectors.git.description') }}</p>

                            <div v-if="loading" class="d-flex align-items-center gap-2 text-secondary">
                                <span class="spinner-border spinner-border-sm" role="status"></span>
                                <span>…</span>
                            </div>

                            <template v-else>
                                <div class="row g-3">
                                    <div class="col-12 col-md-4">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.provider') }}</label>
                                        <select class="form-select" v-model="form.provider" :disabled="!available">
                                            <option v-for="provider in PROVIDERS" :key="provider.id" :value="provider.id">
                                                {{ provider.label }}
                                            </option>
                                        </select>
                                    </div>
                                    <div class="col-12 col-md-8">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.url') }}</label>
                                        <input type="url" class="form-control" v-model="form.url" :disabled="!available"
                                            :placeholder="$t('admin.connectors.git.form.urlPlaceholder')">
                                    </div>
                                    <div class="col-12 col-md-4">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.branch') }}</label>
                                        <input type="text" class="form-control" v-model="form.branch" :disabled="!available"
                                            placeholder="main">
                                    </div>
                                    <div class="col-12 col-md-4">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.directory') }}</label>
                                        <input type="text" class="form-control" v-model="form.directory" :disabled="!available"
                                            placeholder="datalchemist">
                                        <div class="form-text">{{ $t('admin.connectors.git.form.directoryHelp') }}</div>
                                    </div>
                                    <div class="col-12 col-md-4">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.pollInterval') }}</label>
                                        <input type="number" min="10" step="10" class="form-control" v-model="form.poll_interval"
                                            :disabled="!available">
                                    </div>

                                    <div class="col-12 col-md-4">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.username') }}</label>
                                        <input type="text" class="form-control" v-model="form.username" :disabled="!available"
                                            placeholder="git">
                                        <div class="form-text">{{ $t('admin.connectors.git.form.usernameHelp') }}</div>
                                    </div>
                                    <div class="col-12 col-md-8">
                                        <label class="form-label d-flex align-items-center gap-2">
                                            {{ $t('admin.connectors.git.form.token') }}
                                            <span class="badge rounded-pill"
                                                :class="settings?.has_token ? 'text-bg-success' : 'text-bg-secondary'">
                                                {{ settings?.has_token ? $t('admin.connectors.git.form.set') : $t('admin.connectors.git.form.notSet') }}
                                            </span>
                                        </label>
                                        <div class="input-group">
                                            <input type="password" class="form-control" v-model="credentials.token"
                                                autocomplete="new-password" :disabled="!available || clear.token"
                                                :placeholder="settings?.has_token ? $t('admin.connectors.git.form.unchanged') : ''">
                                            <button v-if="settings?.has_token" type="button" class="btn btn-outline-secondary"
                                                :class="{ active: clear.token }" @click="clear.token = !clear.token">
                                                {{ $t('admin.connectors.git.form.clear') }}
                                            </button>
                                        </div>
                                    </div>

                                    <div class="col-12 col-md-6">
                                        <label class="form-label d-flex align-items-center gap-2">
                                            {{ $t('admin.connectors.git.form.passphrase') }}
                                            <span class="badge rounded-pill"
                                                :class="settings?.has_passphrase ? 'text-bg-success' : 'text-bg-secondary'">
                                                {{ settings?.has_passphrase ? $t('admin.connectors.git.form.set') : $t('admin.connectors.git.form.notSet') }}
                                            </span>
                                        </label>
                                        <div class="input-group">
                                            <input type="password" class="form-control" v-model="credentials.passphrase"
                                                autocomplete="new-password" :disabled="!available || clear.passphrase"
                                                :placeholder="settings?.has_passphrase ? $t('admin.connectors.git.form.unchanged') : ''">
                                            <button v-if="settings?.has_passphrase" type="button" class="btn btn-outline-secondary"
                                                :class="{ active: clear.passphrase }" @click="clear.passphrase = !clear.passphrase">
                                                {{ $t('admin.connectors.git.form.clear') }}
                                            </button>
                                        </div>
                                        <div class="form-text">{{ $t('admin.connectors.git.form.passphraseHelp') }}</div>
                                    </div>
                                    <div class="col-12 col-md-6">
                                        <label class="form-label d-flex align-items-center gap-2">
                                            {{ $t('admin.connectors.git.form.webhookSecret') }}
                                            <span class="badge rounded-pill"
                                                :class="settings?.has_webhook_secret ? 'text-bg-success' : 'text-bg-secondary'">
                                                {{ settings?.has_webhook_secret ? $t('admin.connectors.git.form.set') : $t('admin.connectors.git.form.notSet') }}
                                            </span>
                                        </label>
                                        <div class="input-group">
                                            <input type="password" class="form-control" v-model="credentials.webhook_secret"
                                                autocomplete="new-password" :disabled="!available || clear.webhook_secret"
                                                :placeholder="settings?.has_webhook_secret ? $t('admin.connectors.git.form.unchanged') : ''">
                                            <button v-if="settings?.has_webhook_secret" type="button"
                                                class="btn btn-outline-secondary" :class="{ active: clear.webhook_secret }"
                                                @click="clear.webhook_secret = !clear.webhook_secret">
                                                {{ $t('admin.connectors.git.form.clear') }}
                                            </button>
                                        </div>
                                        <div class="form-text">
                                            {{ $t('admin.connectors.git.form.webhookHelp') }}
                                            <code class="admin-connectors-webhook-url" role="button" :title="webhookUrl"
                                                @click="copyWebhookUrl">{{ webhookUrl }}</code>
                                        </div>
                                    </div>

                                    <div class="col-12">
                                        <label class="form-label">{{ $t('admin.connectors.git.form.author') }}</label>
                                        <div class="row g-2">
                                            <div class="col-12 col-md-6">
                                                <input type="text" class="form-control" v-model="form.author_name"
                                                    :disabled="!available" :placeholder="$t('admin.connectors.git.form.authorName')">
                                            </div>
                                            <div class="col-12 col-md-6">
                                                <input type="email" class="form-control" v-model="form.author_email"
                                                    :disabled="!available" :placeholder="$t('admin.connectors.git.form.authorEmail')">
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div v-if="testResult" class="alert mb-0 py-2 px-3 small"
                                    :class="testResult.reachable ? (testResult.branch_found || testResult.empty ? 'alert-success' : 'alert-warning') : 'alert-danger'">
                                    <template v-if="!testResult.reachable">
                                        <i class="bi bi-x-octagon-fill me-1"></i>{{ testResult.error }}
                                    </template>
                                    <template v-else-if="testResult.empty">
                                        <i class="bi bi-check-circle-fill me-1"></i>{{ $t('admin.connectors.git.test.ok') }}
                                        — {{ $t('admin.connectors.git.test.empty') }}
                                    </template>
                                    <template v-else-if="testResult.branch_found">
                                        <i class="bi bi-check-circle-fill me-1"></i>{{ $t('admin.connectors.git.test.ok') }}
                                        — {{ $t('admin.connectors.git.test.branchFound', { branch: form.branch || 'main' }) }}
                                    </template>
                                    <template v-else>
                                        <i class="bi bi-exclamation-triangle-fill me-1"></i>
                                        {{ $t('admin.connectors.git.test.branchMissing', { branch: form.branch || 'main', branches: (testResult.branches || []).join(', ') || '—' }) }}
                                    </template>
                                </div>

                                <div class="d-flex flex-wrap align-items-center gap-2 mt-auto">
                                    <button type="button" class="btn btn-outline-secondary" :disabled="!available || testing || !form.url"
                                        @click="testConnection">
                                        <span v-if="testing" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                        <i v-else class="bi bi-wifi me-1"></i>
                                        {{ $t('admin.connectors.git.actions.test') }}
                                    </button>
                                    <button type="button" class="btn btn-primary" :disabled="!available || saving || !dirty"
                                        @click="saveSettings">
                                        <span v-if="saving" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                        <i v-else class="bi bi-save me-1"></i>
                                        {{ $t('admin.connectors.git.actions.save') }}
                                    </button>
                                    <span v-if="notice === 'saved'" class="small text-success">
                                        <i class="bi bi-check2 me-1"></i>{{ $t('admin.connectors.git.saved') }}
                                    </span>
                                    <span v-else-if="notice === 'copied'" class="small text-success">
                                        <i class="bi bi-clipboard-check me-1"></i>{{ $t('admin.connectors.git.copied') }}
                                    </span>
                                </div>
                            </template>
                        </div>
                    </article>
                </div>

                <!-- État -->
                <div class="col-12 col-xl-5">
                    <article class="card admin-connectors-panel shadow-sm h-100">
                        <div class="card-body p-4 d-flex flex-column gap-3">
                            <h5 class="admin-connectors-panel-title mb-0">{{ $t('admin.connectors.git.status.title') }}</h5>

                            <dl class="admin-connectors-status mb-0">
                                <dt>{{ $t('admin.connectors.git.status.lastSync') }}</dt>
                                <dd>{{ formatDate(status?.last_sync_at) || $t('admin.connectors.git.status.never') }}</dd>
                                <dt>{{ $t('admin.connectors.git.status.lastCommit') }}</dt>
                                <dd>
                                    <code v-if="status?.last_commit" :title="status.last_commit">{{ shortCommit(status.last_commit) }}</code>
                                    <span v-else>—</span>
                                    <span v-if="status?.last_sync_at" class="text-secondary small ms-2">
                                        {{ status.last_pushed }} {{ $t('admin.connectors.git.status.pushed') }},
                                        {{ status.last_pulled }} {{ $t('admin.connectors.git.status.pulled') }}
                                    </span>
                                </dd>
                                <template v-if="status?.last_error">
                                    <dt class="text-danger">{{ $t('admin.connectors.git.status.lastError') }}</dt>
                                    <dd class="text-danger">
                                        {{ status.last_error }}
                                        <span class="text-secondary small d-block">{{ formatDate(status.last_error_at) }}</span>
                                    </dd>
                                </template>
                            </dl>

                            <div v-if="warnings.length" class="alert alert-warning mb-0 py-2 px-3 small">
                                <p class="fw-semibold mb-1">{{ $t('admin.connectors.git.status.warnings') }}</p>
                                <ul class="mb-0 ps-3">
                                    <li v-for="warning in warnings" :key="warning">{{ warning }}</li>
                                </ul>
                            </div>

                            <div class="d-flex flex-wrap gap-2 mt-auto">
                                <template v-if="!enabled">
                                    <button type="button" class="btn btn-success" :disabled="!available || enabling || dirty || !form.url"
                                        :title="dirty ? $t('admin.connectors.git.status.saveFirst') : ''" @click="requestEnable">
                                        <span v-if="enabling" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                        <i v-else class="bi bi-play-circle-fill me-1"></i>
                                        {{ $t('admin.connectors.git.actions.enable') }}
                                    </button>
                                    <span v-if="dirty" class="small text-secondary align-self-center">
                                        {{ $t('admin.connectors.git.status.saveFirst') }}
                                    </span>
                                </template>
                                <template v-else>
                                    <button type="button" class="btn btn-primary" :disabled="syncing || status?.running" @click="syncNow">
                                        <span v-if="syncing || status?.running" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                        <i v-else class="bi bi-arrow-repeat me-1"></i>
                                        {{ $t('admin.connectors.git.actions.sync') }}
                                    </button>
                                    <button type="button" class="btn btn-outline-danger" :disabled="disabling" @click="disableModal = true">
                                        <i class="bi bi-stop-circle me-1"></i>
                                        {{ $t('admin.connectors.git.actions.disable') }}
                                    </button>
                                </template>
                            </div>
                        </div>
                    </article>
                </div>

                <!-- Conflits -->
                <div class="col-12">
                    <article class="card admin-connectors-panel admin-connectors-table-panel shadow-sm">
                        <div class="card-body p-0 d-flex flex-column">
                            <div class="admin-connectors-panel-head px-4 py-3 d-flex align-items-center gap-2">
                                <i class="bi bi-exclamation-triangle-fill" :class="conflicts.length ? 'text-warning' : 'text-secondary'"></i>
                                <h5 class="admin-connectors-panel-title mb-0">{{ $t('admin.connectors.git.conflicts.title') }}</h5>
                                <span class="badge rounded-pill ms-1" :class="conflicts.length ? 'text-bg-warning' : 'text-bg-secondary'">
                                    {{ conflicts.length }}
                                </span>
                                <span class="text-secondary small ms-auto d-none d-md-inline">
                                    {{ $t('admin.connectors.git.conflicts.help') }}
                                </span>
                            </div>

                            <p v-if="conflicts.length === 0" class="text-secondary px-4 py-3 mb-0">
                                {{ $t('admin.connectors.git.conflicts.empty') }}
                            </p>

                            <div v-else class="admin-connectors-table-wrap">
                                <table class="table align-middle mb-0 admin-connectors-table">
                                    <thead>
                                        <tr>
                                            <th scope="col">{{ $t('admin.connectors.git.conflicts.entity') }}</th>
                                            <th scope="col">{{ $t('admin.connectors.git.conflicts.reason') }}</th>
                                            <th scope="col" class="text-end">{{ $t('admin.users.actions') }}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr v-for="conflict in conflicts" :key="`${conflict.kind}-${conflict.id}`">
                                            <td>
                                                <div class="d-flex align-items-center gap-2">
                                                    <i :class="kindIcon(conflict.kind)"></i>
                                                    <span class="text-secondary small">{{ $t(`admin.connectors.git.kinds.${conflict.kind}`) }}</span>
                                                    <RouterLink v-if="!conflict.local_deleted" :to="editRoute(conflict)" class="fw-semibold text-decoration-none">
                                                        {{ conflict.name }}
                                                    </RouterLink>
                                                    <span v-else class="fw-semibold">{{ conflict.name }}</span>
                                                    <span class="badge rounded-pill bg-light text-dark border font-monospace">#{{ conflict.id }}</span>
                                                </div>
                                            </td>
                                            <td class="text-secondary">{{ conflict.reason }}</td>
                                            <td class="text-end">
                                                <div class="btn-group btn-group-sm">
                                                    <button type="button" class="btn btn-outline-secondary" @click="openConflict(conflict)">
                                                        <i class="bi bi-arrow-left-right me-1"></i>{{ $t('admin.connectors.git.actions.compare') }}
                                                    </button>
                                                    <button type="button" class="btn btn-outline-primary" :disabled="Boolean(resolving)"
                                                        @click="resolve(conflict, 'local')">
                                                        <span v-if="isResolving(conflict, 'local')" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                                        <i v-else class="bi bi-hdd-fill me-1"></i>{{ $t('admin.connectors.git.actions.keepLocal') }}
                                                    </button>
                                                    <button type="button" class="btn btn-outline-warning" :disabled="Boolean(resolving)"
                                                        @click="resolve(conflict, 'remote')">
                                                        <span v-if="isResolving(conflict, 'remote')" class="spinner-border spinner-border-sm me-1" role="status"></span>
                                                        <i v-else class="bi bi-git me-1"></i>{{ $t('admin.connectors.git.actions.keepRemote') }}
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </article>
                </div>
            </div>
        </div>

        <!-- Modale : direction du premier alignement -->
        <div v-if="enableModal" class="modal fade show d-block admin-connectors-modal" tabindex="-1" role="dialog" aria-modal="true"
            @click.self="enableModal = null">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content shadow-lg border-0">
                    <div class="modal-header border-0 bg-body-tertiary">
                        <h5 class="modal-title d-flex align-items-center gap-2">
                            <i class="bi bi-signpost-split-fill"></i>
                            <span>{{ $t('admin.connectors.git.enableModal.title') }}</span>
                        </h5>
                        <button type="button" class="btn-close" aria-label="Close" @click="enableModal = null"></button>
                    </div>
                    <div class="modal-body d-flex flex-column gap-3">
                        <p class="mb-0">{{ $t('admin.connectors.git.enableModal.intro') }}</p>
                        <div class="list-group admin-connectors-choices">
                            <label v-for="choice in ['merge', 'local', 'remote']" :key="choice"
                                class="list-group-item list-group-item-action d-flex gap-3 align-items-start"
                                :class="{ active: enableModal.direction === choice }">
                                <input class="form-check-input mt-1 flex-shrink-0" type="radio" :value="choice" v-model="enableModal.direction">
                                <span>
                                    <span class="fw-semibold d-block">{{ $t(`admin.connectors.git.enableModal.${choice}`) }}</span>
                                    <span class="small">{{ $t(`admin.connectors.git.enableModal.${choice}Help`) }}</span>
                                </span>
                            </label>
                        </div>
                    </div>
                    <div class="modal-footer border-0">
                        <button type="button" class="btn btn-outline-secondary" @click="enableModal = null">
                            {{ $t('admin.connectors.git.actions.cancel') }}
                        </button>
                        <button type="button" class="btn btn-success" :disabled="enabling" @click="confirmEnable">
                            <span v-if="enabling" class="spinner-border spinner-border-sm me-1" role="status"></span>
                            <i v-else class="bi bi-play-circle-fill me-1"></i>
                            {{ $t('admin.connectors.git.actions.enable') }}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Modale : désactivation -->
        <div v-if="disableModal" class="modal fade show d-block admin-connectors-modal" tabindex="-1" role="dialog" aria-modal="true"
            @click.self="disableModal = false">
            <div class="modal-dialog modal-dialog-centered" role="document">
                <div class="modal-content shadow-lg border-0">
                    <div class="modal-header border-0 bg-body-tertiary">
                        <h5 class="modal-title d-flex align-items-center gap-2">
                            <i class="bi bi-stop-circle text-danger"></i>
                            <span>{{ $t('admin.connectors.git.disableModal.title') }}</span>
                        </h5>
                        <button type="button" class="btn-close" aria-label="Close" @click="disableModal = false"></button>
                    </div>
                    <div class="modal-body">
                        <p class="mb-0">{{ $t('admin.connectors.git.disableModal.body') }}</p>
                    </div>
                    <div class="modal-footer border-0">
                        <button type="button" class="btn btn-outline-secondary" @click="disableModal = false">
                            {{ $t('admin.connectors.git.actions.cancel') }}
                        </button>
                        <button type="button" class="btn btn-danger" :disabled="disabling" @click="disable">
                            <span v-if="disabling" class="spinner-border spinner-border-sm me-1" role="status"></span>
                            {{ $t('admin.connectors.git.actions.disable') }}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Modale : comparaison d'un conflit -->
        <div v-if="conflictModal" class="modal fade show d-block admin-connectors-modal" tabindex="-1" role="dialog" aria-modal="true"
            @click.self="conflictModal = null">
            <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable modal-xl" role="document">
                <div class="modal-content shadow-lg border-0">
                    <div class="modal-header border-0 bg-body-tertiary">
                        <h5 class="modal-title d-flex align-items-center gap-2 flex-wrap">
                            <i :class="kindIcon(conflictModal.conflict.kind)"></i>
                            <span>{{ conflictModal.conflict.name }}</span>
                            <span class="badge rounded-pill bg-light text-dark border font-monospace">#{{ conflictModal.conflict.id }}</span>
                            <span class="small text-secondary fw-normal">{{ conflictModal.conflict.reason }}</span>
                        </h5>
                        <button type="button" class="btn-close" aria-label="Close" @click="conflictModal = null"></button>
                    </div>
                    <div class="modal-body d-flex flex-column gap-3">
                        <div v-for="name in fileNames(conflictModal)" :key="name" class="admin-connectors-diff">
                            <div class="admin-connectors-diff-name font-monospace small">{{ name }}</div>
                            <div class="row g-2">
                                <div class="col-12 col-lg-6">
                                    <div class="admin-connectors-diff-label"><i class="bi bi-hdd-fill me-1"></i>{{ $t('admin.connectors.git.conflicts.localVersion') }}</div>
                                    <pre v-if="conflictModal.local?.[name] !== undefined" class="admin-connectors-diff-body">{{ conflictModal.local[name] }}</pre>
                                    <p v-else class="text-secondary small fst-italic mb-0">{{ $t('admin.connectors.git.conflicts.deleted') }}</p>
                                </div>
                                <div class="col-12 col-lg-6">
                                    <div class="admin-connectors-diff-label"><i class="bi bi-git me-1"></i>{{ $t('admin.connectors.git.conflicts.remoteVersion') }}</div>
                                    <pre v-if="conflictModal.remote?.[name] !== undefined" class="admin-connectors-diff-body">{{ conflictModal.remote[name] }}</pre>
                                    <p v-else class="text-secondary small fst-italic mb-0">{{ $t('admin.connectors.git.conflicts.deleted') }}</p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="modal-footer border-0">
                        <button type="button" class="btn btn-outline-secondary me-auto" @click="conflictModal = null">
                            {{ $t('admin.connectors.git.actions.close') }}
                        </button>
                        <button type="button" class="btn btn-primary" :disabled="Boolean(resolving)"
                            @click="resolve(conflictModal.conflict, 'local')">
                            <span v-if="isResolving(conflictModal.conflict, 'local')" class="spinner-border spinner-border-sm me-1" role="status"></span>
                            <i v-else class="bi bi-hdd-fill me-1"></i>{{ $t('admin.connectors.git.actions.keepLocal') }}
                        </button>
                        <button type="button" class="btn btn-warning" :disabled="Boolean(resolving)"
                            @click="resolve(conflictModal.conflict, 'remote')">
                            <span v-if="isResolving(conflictModal.conflict, 'remote')" class="spinner-border spinner-border-sm me-1" role="status"></span>
                            <i v-else class="bi bi-git me-1"></i>{{ $t('admin.connectors.git.actions.keepRemote') }}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="enableModal || disableModal || conflictModal" class="modal-backdrop fade show" style="z-index: 1040;"></div>
    </section>
</template>
