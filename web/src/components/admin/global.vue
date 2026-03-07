<script setup>
import { computed, inject, onMounted, ref, watch } from 'vue';
import axios from 'axios';

const apiUrl = inject('apiUrl');
const save = inject('save');

const parameter = ref(null);
const originalParameter = ref(null);
const views = ref([]);

const languageOptions = [
    { id: 'en', label: 'English', flagCode: 'gb' },
    { id: 'fr', label: 'Français', flagCode: 'fr' },
];

save.value.safe();

const hasUnsavedChanges = computed(() => {
    if (!parameter.value || !originalParameter.value) {
        return false;
    }

    return Object.keys(parameter.value).some(
        (key) => `${parameter.value[key]}` !== `${originalParameter.value[key]}`
    );
});

const isLdapEnabled = computed(() => parameter.value?.ldap === true || parameter.value?.ldap === 'true');
const isThemeLight = computed(() => parameter.value?.theme === 'light');
const isThemeDark = computed(() => parameter.value?.theme === 'dark');

function cloneParameters(data) {
    return JSON.parse(JSON.stringify(data));
}

async function saveParameters() {
    if (!parameter.value || !originalParameter.value) {
        return;
    }

    const changedEntries = Object.entries(parameter.value).filter(
        ([key, value]) => `${value}` !== `${originalParameter.value[key]}`
    );

    if (changedEntries.length === 0) {
        save.value.status.show();
        return;
    }

    try {
        await Promise.all(
            changedEntries.map(([key, value]) =>
                axios.put(`${apiUrl}/parameter/${key}`, {
                    Name: key,
                    Value: `${value}`,
                })
            )
        );

        localStorage.setItem('reloadparameters', true);
        originalParameter.value = cloneParameters(parameter.value);
        save.value.status.show();
    } catch (error) {
        console.error('Failed to save global parameters', error);
        save.value.status.error();
    }
}

async function fetchParameters() {
    try {
        const response = await axios.get(`${apiUrl}/parameters/admin`);
        parameter.value = response.data;
        originalParameter.value = cloneParameters(response.data);
    } catch (error) {
        console.error('Failed to fetch admin parameters', error);
    }
}

async function fetchViews() {
    try {
        const response = await axios.get(`${apiUrl}/views`);
        views.value = response.data || [];
    } catch (error) {
        console.error('Failed to fetch views', error);
    }
}

watch(hasUnsavedChanges, (dirty) => {
    if (dirty) {
        save.value.status.saveable();
    } else {
        save.value.status.show();
    }
});

onMounted(async () => {
    save.value.function = saveParameters;
    save.value.status.show();
    await Promise.all([fetchParameters(), fetchViews()]);
});
</script>

<template>
    <section class="admin-global-page container-fluid px-0 py-1 py-lg-2">
        <div v-if="parameter" class="d-flex flex-column gap-4">
            <header class="card admin-hero shadow-sm">
                <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
                    <div class="admin-hero-icon">
                        <i class="bi bi-sliders2-vertical"></i>
                    </div>
                    <div class="flex-grow-1">
                        <p class="admin-kicker mb-1">{{ $t('admin.header') }}</p>
                        <h4 class="mb-1">{{ $t('admin.global.header') }}</h4>
                        <p class="mb-0 text-secondary">{{ parameter.name }}</p>
                    </div>
                    <span class="badge rounded-pill admin-state-chip"
                        :class="hasUnsavedChanges ? 'text-bg-warning' : 'text-bg-success'">
                        <i
                            :class="hasUnsavedChanges ? 'bi bi-exclamation-circle-fill me-1' : 'bi bi-check-circle-fill me-1'"></i>
                        {{ hasUnsavedChanges ? $t('admin.global.pending') : $t('admin.global.synced') }}
                    </span>
                </div>
            </header>

            <div class="row g-4">
                <div class="col-12 col-xl-5">
                    <article class="card admin-panel shadow-sm h-100">
                        <div class="card-body p-4">
                            <h5 class="panel-title">{{ $t('admin.global.header') }}</h5>

                            <div class="mb-3">
                                <label class="form-label">{{ $t('admin.global.name') }}</label>
                                <div class="input-group">
                                    <span class="input-group-text"><i class="bi bi-card-heading"></i></span>
                                    <input type="text" class="form-control" v-model="parameter.name">
                                </div>
                            </div>

                            <div class="mb-0">
                                <label class="form-label">{{ $t('admin.global.defaultview') }}</label>
                                <div class="input-group">
                                    <span class="input-group-text"><i class="bi bi-house-door"></i></span>
                                    <select class="form-select" v-model="parameter.defaultview">
                                        <option value="">-</option>
                                        <option v-for="view in views" :key="view.id" :value="view.id">{{ view.name }}
                                        </option>
                                    </select>
                                </div>
                            </div>

                        </div>
                    </article>
                </div>

                <div class="col-12 col-xl-7">
                    <article class="card admin-panel shadow-sm h-100">
                        <div class="card-body p-4">
                            <h5 class="panel-title">{{ $t('admin.global.bg') }}</h5>

                            <div class="row g-3">
                                <div class="col-md-6">
                                    <label class="form-label d-block">{{ $t('admin.global.theme') }}</label>
                                    <div class="theme-toggle"
                                        :class="{ 'is-light': isThemeLight, 'is-dark': isThemeDark }" role="group"
                                        :aria-label="$t('admin.global.theme')">
                                        <span class="theme-toggle-glider" aria-hidden="true"></span>
                                        <button type="button" class="theme-toggle-btn"
                                            :class="{ 'is-active': isThemeLight }" :aria-pressed="isThemeLight"
                                            @click="parameter.theme = 'light'">
                                            <i class="bi bi-brightness-high-fill"></i>
                                            <span>{{ $t('admin.global.light') }}</span>
                                        </button>
                                        <button type="button" class="theme-toggle-btn"
                                            :class="{ 'is-active': isThemeDark }" :aria-pressed="isThemeDark"
                                            @click="parameter.theme = 'dark'">
                                            <i class="bi bi-moon-stars-fill"></i>
                                            <span>{{ $t('admin.global.dark') }}</span>
                                        </button>
                                    </div>
                                </div>

                                <div class="col-md-6">
                                    <label class="form-label d-block">{{ $t('admin.global.lang') }}</label>
                                    <div class="language-toggle" role="group" :aria-label="$t('admin.global.lang')">
                                        <button v-for="language in languageOptions" :key="language.id" type="button"
                                            class="language-toggle-btn"
                                            :class="{ 'is-active': parameter.lang === language.id }"
                                            :aria-pressed="parameter.lang === language.id"
                                            @click="parameter.lang = language.id">
                                            <span class="lang-flag fi" :class="`fi-${language.flagCode}`"
                                                aria-hidden="true"></span>
                                            <span>{{ language.label }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <div class="row g-3 mt-1">
                                <div class="col-md-6">
                                    <div class="color-card">
                                        <div class="color-card-title">
                                            <i class="bi bi-sun-fill"></i>
                                            <span>{{ $t('admin.global.light') }}</span>
                                        </div>
                                        <div class="color-row">
                                            <label class="color-input-wrap">
                                                <span>A</span>
                                                <input type="color" v-model="parameter.bg_color_light"
                                                    class="form-control form-control-color">
                                            </label>
                                            <label class="color-input-wrap">
                                                <span>B</span>
                                                <input type="color" v-model="parameter.bg_color2_light"
                                                    class="form-control form-control-color">
                                            </label>
                                        </div>
                                        <small class="text-secondary color-value">
                                            {{ parameter.bg_color_light }} - {{ parameter.bg_color2_light }}
                                        </small>
                                    </div>
                                </div>

                                <div class="col-md-6">
                                    <div class="color-card">
                                        <div class="color-card-title">
                                            <i class="bi bi-moon-stars-fill"></i>
                                            <span>{{ $t('admin.global.dark') }}</span>
                                        </div>
                                        <div class="color-row">
                                            <label class="color-input-wrap">
                                                <span>A</span>
                                                <input type="color" v-model="parameter.bg_color_dark"
                                                    class="form-control form-control-color">
                                            </label>
                                            <label class="color-input-wrap">
                                                <span>B</span>
                                                <input type="color" v-model="parameter.bg_color2_dark"
                                                    class="form-control form-control-color">
                                            </label>
                                        </div>
                                        <small class="text-secondary color-value">
                                            {{ parameter.bg_color_dark }} - {{ parameter.bg_color2_dark }}
                                        </small>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </article>
                </div>

                <div class="col-12">
                    <article class="card admin-panel shadow-sm">
                        <div class="card-body p-4">
                            <h5 class="panel-title">{{ $t('admin.global.ldap.header') }}</h5>

                            <div class="row g-3 mb-3">
                                <div class="col-md-4">
                                    <div class="form-check form-switch">
                                        <input class="form-check-input" type="checkbox" id="ldap-enable"
                                            v-model="parameter.ldap">
                                        <label class="form-check-label" for="ldap-enable">{{
                                            $t('admin.global.ldap.enable') }}</label>
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="form-check form-switch">
                                        <input class="form-check-input" type="checkbox" id="ldap-ssl"
                                            v-model="parameter.ldap_ssl">
                                        <label class="form-check-label" for="ldap-ssl">SSL (LDAPS)</label>
                                    </div>
                                </div>
                                <div class="col-md-4">
                                    <div class="form-check form-switch">
                                        <input class="form-check-input" type="checkbox" id="ldap-skip-verify"
                                            v-model="parameter.ldap_skip_verify">
                                        <label class="form-check-label" for="ldap-skip-verify">{{
                                            $t('admin.global.ldap.skipverify') }}</label>
                                    </div>
                                </div>
                            </div>

                            <div class="row g-3">
                                <div class="col-md-3">
                                    <label class="form-label">{{ $t('admin.global.ldap.server') }}</label>
                                    <input type="text" class="form-control" v-model="parameter.ldap_host"
                                        :disabled="!isLdapEnabled">
                                </div>
                                <div class="col-md-2">
                                    <label class="form-label">{{ $t('admin.global.ldap.port') }}</label>
                                    <input type="number" class="form-control" placeholder="389 / 636" min="0"
                                        v-model="parameter.ldap_port" :disabled="!isLdapEnabled">
                                </div>
                                <div class="col-md-3">
                                    <label class="form-label">{{ $t('admin.global.ldap.basedn') }}</label>
                                    <input type="text" class="form-control" v-model="parameter.ldap_base_dn"
                                        :disabled="!isLdapEnabled">
                                </div>
                                <div class="col-md-4">
                                    <label class="form-label">{{ $t('admin.global.ldap.login') }}</label>
                                    <input type="text" class="form-control" placeholder="uid, samaccountname, mail, ..."
                                        v-model="parameter.ldap_filter" :disabled="!isLdapEnabled">
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">{{ $t('admin.global.ldap.user') }}</label>
                                    <input type="text" class="form-control" v-model="parameter.ldap_user"
                                        :disabled="!isLdapEnabled">
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">{{ $t('admin.global.ldap.password') }}</label>
                                    <input type="password" class="form-control" v-model="parameter.ldap_password"
                                        :disabled="!isLdapEnabled">
                                </div>
                            </div>
                        </div>
                    </article>
                </div>
            </div>
        </div>

        <div v-else class="card admin-panel shadow-sm">
            <div class="card-body p-4">
                <div class="placeholder-glow">
                    <span class="placeholder col-4 mb-3"></span>
                    <span class="placeholder col-8 mb-2"></span>
                    <span class="placeholder col-7 mb-2"></span>
                    <span class="placeholder col-6"></span>
                </div>
            </div>
        </div>
    </section>
</template>