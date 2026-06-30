<script setup>
import { computed, ref, onMounted, onBeforeUnmount, inject, watch } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');
const parameter = inject('parameters');

const user = ref(null)
const isSaving = ref(false)
const passwordConfirm = ref('')
const saveSuccessVisible = ref(false)
let saveSuccessTimer = null

const languageOptions = [
    { id: 'default', label: 'Default' },
    { id: 'en', label: 'English', flagCode: 'gb' },
    { id: 'fr', label: 'Français', flagCode: 'fr' },
]

const themeOptions = [
    { id: 'default', labelKey: 'profil.defaultPreview', icon: 'bi bi-display' },
    { id: 'light', labelKey: 'admin.global.light', icon: 'bi bi-brightness-high-fill' },
    { id: 'dark', labelKey: 'admin.global.dark', icon: 'bi bi-moon-stars-fill' },
]

const userInitial = computed(() => {
    const firstChar = user.value?.name?.trim()?.charAt(0)
    return firstChar ? firstChar.toUpperCase() : '?'
})

const currentLanguageOption = computed(() => {
    if (!user.value) {
        return null
    }
    return languageOptions.find((item) => item.id === user.value.lang) || null
})

const currentThemeOption = computed(() => {
    if (!user.value) {
        return null
    }
    return themeOptions.find((item) => item.id === user.value.theme) || null
})

const passwordMismatch = computed(() => {
    if (!user.value) {
        return false
    }

    const password = user.value.password || ''
    const confirmation = passwordConfirm.value || ''

    if (!password && !confirmation) {
        return false
    }

    return password !== confirmation
})

const canSave = computed(() => {
    return !!user.value && !passwordMismatch.value && !isSaving.value
})

function buildUserPayload() {
    return {
        id: Number(user.value?.id) || 0,
        name: user.value?.name || '',
        type: user.value?.type || '',
        lang: user.value?.lang || 'default',
        theme: user.value?.theme || 'default',
        password: user.value?.password || '',
    }
}

function clearSaveSuccessTimer() {
    if (saveSuccessTimer !== null) {
        clearTimeout(saveSuccessTimer)
        saveSuccessTimer = null
    }
}

function showSaveSuccessNotice() {
    clearSaveSuccessTimer()
    saveSuccessVisible.value = true
    saveSuccessTimer = setTimeout(() => {
        saveSuccessVisible.value = false
        saveSuccessTimer = null
    }, 2600)
}

const fetchUser = async () => {
    try {
        const response = await axios.get(`${apiUrl}/user`)
        user.value = response.data
    } catch (error) {
        console.error(`Erreur lors de la récupération de l'utilisateur`, error)
    }
};

async function UpdateUser() {
    if (!canSave.value) {
        return
    }

    isSaving.value = true
    try {
        const payload = buildUserPayload()
        await axios.put(`${apiUrl}/user`, payload)
        await fetchUser()
        if (user.value) {
            user.value.password = ''
        }
        passwordConfirm.value = ''
        showSaveSuccessNotice()
    } catch (error) {
        saveSuccessVisible.value = false
        console.error('Erreur lors de la mise a jour du profil', error?.response?.data || error)
    } finally {
        isSaving.value = false
    }
}

watch(user, () => {
    if (!user.value) {
        return
    }

    if (user.value.theme !== "default") {
        parameter.value.theme = user.value.theme
    }
    if (user.value.lang !== "default") {
        parameter.value.lang = user.value.lang
    }
}, { deep: true });

watch(parameter, () => {
    localStorage.setItem('reloadparameters', true);
}, { deep: true });

onMounted(() => {
    fetchUser();
});

onBeforeUnmount(() => {
    clearSaveSuccessTimer()
});
</script>

<template>
    <section class="profile-page container py-2 py-lg-4">
        <div v-if="user">
            <header class="card profile-hero border-0 shadow-sm mb-4">
                <div class="card-body d-flex flex-column flex-md-row align-items-md-center gap-3 gap-lg-4">
                    <div class="profile-avatar">{{ userInitial }}</div>
                    <div class="flex-grow-1">
                        <p class="profile-kicker mb-1">{{ $t('profil.header') }}</p>
                        <h4 class="mb-1">{{ user.name }}</h4>
                        <p class="mb-0 text-secondary">{{ $t('profil.parameters') }}</p>
                    </div>
                    <div class="d-flex flex-wrap gap-2">
                        <span class="badge rounded-pill text-bg-light profile-pill">
                            <span v-if="currentLanguageOption?.flagCode" class="lang-flag me-1 fi"
                                :class="`fi-${currentLanguageOption.flagCode}`" aria-hidden="true"></span>
                            <i v-else class="bi bi-globe2 me-1" aria-hidden="true"></i>
                            {{ currentLanguageOption ? currentLanguageOption.label : user.lang }}
                        </span>
                        <span class="badge rounded-pill text-bg-light profile-pill">
                            <i :class="[currentThemeOption?.icon || 'bi bi-circle-half', 'me-1']"></i>
                            {{ currentThemeOption ? $t(currentThemeOption.labelKey) : user.theme }}
                        </span>
                    </div>
                </div>
            </header>

            <div class="row g-4">
                <div class="col-12">
                    <article class="card profile-editor border-0 shadow-sm">
                        <div class="card-body p-4 p-xl-5">
                            <h5 class="card-title mb-3">{{ $t('profil.parameters') }}</h5>

                            <div class="row g-3">
                                <div class="col-md-6">
                                    <label class="form-label d-block">{{ $t('profil.language') }}</label>
                                    <div class="language-toggle" role="group" :aria-label="$t('profil.language')">
                                        <button v-for="l in languageOptions" :key="l.id" type="button"
                                            class="language-toggle-btn" :class="{ 'is-active': user.lang === l.id }"
                                            :aria-pressed="user.lang === l.id" @click="user.lang = l.id">
                                            <span v-if="l.flagCode" class="lang-flag fi" :class="`fi-${l.flagCode}`"
                                                aria-hidden="true"></span>
                                            <i v-else class="bi bi-globe2" aria-hidden="true"></i>
                                            <span>{{ l.label }}</span>
                                        </button>
                                    </div>
                                </div>

                                <div class="col-md-6">
                                    <label class="form-label d-block">{{ $t('profil.theme') }}</label>
                                    <div class="theme-toggle"
                                        :class="{ 'is-light': user.theme === 'light', 'is-dark': user.theme === 'dark' }"
                                        role="group" :aria-label="$t('profil.theme')">
                                        <span class="theme-toggle-glider" aria-hidden="true"></span>

                                        <button type="button" class="theme-toggle-btn"
                                            :class="{ 'is-active': user.theme === 'light' }"
                                            :aria-pressed="user.theme === 'light'" @click="user.theme = 'light'">
                                            <i class="bi bi-brightness-high-fill"></i>
                                            <span>{{ $t('admin.global.light') }}</span>
                                        </button>

                                        <button type="button" class="theme-toggle-btn"
                                            :class="{ 'is-active': user.theme === 'dark' }"
                                            :aria-pressed="user.theme === 'dark'" @click="user.theme = 'dark'">
                                            <i class="bi bi-moon-stars-fill"></i>
                                            <span>{{ $t('admin.global.dark') }}</span>
                                        </button>
                                    </div>

                                    <button type="button" class="btn profile-default-theme-btn mt-2"
                                        :class="{ 'is-active': user.theme === 'default' }"
                                        @click="user.theme = 'default'">
                                        <i class="bi bi-magic me-1"></i>{{ $t('profil.defaultPreview') }}
                                    </button>
                                </div>
                            </div>

                            <hr class="my-4">

                            <div class="row g-3">
                                <div class="col-md-6">
                                    <label for="inputPassword" class="form-label">{{ $t('profil.password') }}</label>
                                    <div class="input-group has-validation">
                                        <span class="input-group-text"><i class="bi bi-lock"></i></span>
                                        <input type="password" class="form-control" id="inputPassword"
                                            v-model="user.password" :class="{ 'is-invalid': passwordMismatch }">
                                    </div>
                                </div>

                                <div class="col-md-6">
                                    <label for="inputPasswordConfirm" class="form-label">{{ $t('profil.passwordrepeat')
                                    }}</label>
                                    <div class="input-group has-validation">
                                        <span class="input-group-text"><i class="bi bi-lock-fill"></i></span>
                                        <input type="password" class="form-control" id="inputPasswordConfirm"
                                            v-model="passwordConfirm" :class="{ 'is-invalid': passwordMismatch }">
                                    </div>
                                </div>
                            </div>

                            <p v-if="passwordMismatch" class="small text-danger mt-3 mb-0">
                                <i class="bi bi-exclamation-triangle-fill me-1"></i>{{ $t('profil.passwordmismatch') }}
                            </p>

                            <div class="d-flex justify-content-end mt-4">
                                <button type="button" class="btn btn-primary profile-save-btn" :disabled="!canSave"
                                    @click="UpdateUser()">
                                    <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" role="status"
                                        aria-hidden="true"></span>
                                    <i v-else class="bi bi-check2-circle me-2"></i>
                                    {{ $t('profil.save') }}
                                </button>
                            </div>
                        </div>
                    </article>
                </div>
            </div>

        </div>

        <div v-else class="card profile-loading border-0 shadow-sm">
            <div class="card-body p-4">
                <div class="placeholder-glow">
                    <span class="placeholder col-4 mb-3"></span>
                    <span class="placeholder col-8 mb-2"></span>
                    <span class="placeholder col-6 mb-2"></span>
                    <span class="placeholder col-7"></span>
                </div>
            </div>
        </div>

        <div class="profile-toast-wrapper" aria-live="polite" aria-atomic="true">
            <transition name="profile-toast">
                <div v-if="saveSuccessVisible"
                    class="alert alert-success profile-toast d-inline-flex align-items-center gap-2" role="status">
                    <i class="bi bi-check-circle-fill"></i>
                    <span>{{ $t('profil.savesuccess') }}</span>
                </div>
            </transition>
        </div>
    </section>
</template>

<style scoped>
.profile-page {
    max-width: 1080px;
}

.profile-hero,
.profile-editor,
.profile-loading {
    border: 1px solid var(--bs-border-color-translucent);
    border-radius: 1rem;
}

.profile-hero {
    position: relative;
    overflow: hidden;
    background: linear-gradient(120deg, var(--bs-primary-bg-subtle) 0%, var(--bs-body-bg) 55%);
}

.profile-hero::after {
    content: '';
    position: absolute;
    right: -3.25rem;
    top: -3.5rem;
    width: 12.5rem;
    height: 12.5rem;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(var(--bs-primary-rgb), 0.2) 0%, rgba(var(--bs-primary-rgb), 0) 72%);
    pointer-events: none;
}

.profile-avatar {
    width: 3.6rem;
    height: 3.6rem;
    border-radius: 0.9rem;
    display: grid;
    place-items: center;
    color: #fff;
    font-weight: 700;
    font-size: 1.25rem;
    background: linear-gradient(145deg, rgba(var(--bs-primary-rgb), 0.95), rgba(var(--bs-primary-rgb), 0.7));
    box-shadow: 0 0.55rem 1.4rem rgba(var(--bs-primary-rgb), 0.28);
}

.profile-kicker {
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--bs-secondary-color);
}

.profile-pill {
    display: inline-flex;
    align-items: center;
    border: 1px solid var(--bs-border-color-translucent);
    color: var(--bs-emphasis-color);
    font-weight: 500;
}

.profile-editor .form-label {
    font-weight: 600;
    margin-bottom: 0.45rem;
}

.lang-flag {
    width: 1.05rem;
    height: 0.78rem;
    border-radius: 0.2rem;
    border: 1px solid rgba(0, 0, 0, 0.18);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
    flex-shrink: 0;
}

.language-toggle {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.4rem;
}

.language-toggle-btn {
    border: 1px solid var(--bs-border-color);
    background: var(--bs-body-bg);
    color: var(--bs-secondary-color);
    border-radius: 0.75rem;
    min-height: 2.4rem;
    padding: 0.35rem 0.45rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    font-size: 0.9rem;
    font-weight: 600;
    transition: all 0.2s ease;
}

.language-toggle-btn:hover {
    border-color: rgba(var(--bs-primary-rgb), 0.4);
    color: var(--bs-emphasis-color);
}

.language-toggle-btn.is-active {
    border-color: rgba(var(--bs-primary-rgb), 0.55);
    background: rgba(var(--bs-primary-rgb), 0.12);
    color: var(--bs-emphasis-color);
    box-shadow: 0 0 0 0.12rem rgba(var(--bs-primary-rgb), 0.18);
}

.theme-toggle {
    position: relative;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.25rem;
    padding: 0.25rem;
    border: 1px solid var(--bs-border-color);
    border-radius: 0.9rem;
    background: var(--bs-tertiary-bg);
}

.theme-toggle-glider {
    position: absolute;
    top: 0.25rem;
    left: 0.25rem;
    width: calc(50% - 0.375rem);
    height: calc(100% - 0.5rem);
    border-radius: 0.7rem;
    background: var(--bs-body-bg);
    border: 1px solid var(--bs-border-color-translucent);
    box-shadow: 0 0.3rem 0.8rem rgba(0, 0, 0, 0.1);
    transition: transform 0.24s ease, opacity 0.2s ease;
}

.theme-toggle.is-dark .theme-toggle-glider {
    transform: translateX(calc(100% + 0.25rem));
}

.theme-toggle:not(.is-light):not(.is-dark) .theme-toggle-glider {
    opacity: 0;
}

.theme-toggle-btn {
    position: relative;
    z-index: 1;
    border: 0;
    background: transparent;
    color: var(--bs-secondary-color);
    display: inline-flex;
    justify-content: center;
    align-items: center;
    gap: 0.45rem;
    font-weight: 600;
    border-radius: 0.7rem;
    padding: 0.5rem 0.65rem;
    transition: color 0.2s ease;
}

.theme-toggle-btn.is-active {
    color: var(--bs-emphasis-color);
}

.profile-default-theme-btn {
    width: 100%;
    text-align: center;
    border-radius: 0.75rem;
    border: 1px dashed var(--bs-border-color);
    color: var(--bs-secondary-color);
    background: transparent;
}

.profile-default-theme-btn:hover,
.profile-default-theme-btn.is-active {
    border-style: solid;
    border-color: rgba(var(--bs-primary-rgb), 0.45);
    color: var(--bs-emphasis-color);
    background: rgba(var(--bs-primary-rgb), 0.08);
}

.profile-editor .input-group-text {
    background-color: var(--bs-tertiary-bg);
    border-color: var(--bs-border-color);
    color: var(--bs-secondary-color);
}

.profile-editor .form-control,
.profile-editor .form-select {
    border-color: var(--bs-border-color);
}

.profile-editor .form-control:focus,
.profile-editor .form-select:focus {
    border-color: rgba(var(--bs-primary-rgb), 0.45);
    box-shadow: 0 0 0 0.2rem rgba(var(--bs-primary-rgb), 0.15);
}

.profile-save-btn {
    min-width: 12rem;
    border-radius: 0.75rem;
}

.profile-toast-wrapper {
    position: fixed;
    right: 1rem;
    bottom: 1rem;
    z-index: 1090;
    pointer-events: none;
}

.profile-toast {
    margin: 0;
    border-radius: 0.85rem;
    border: 1px solid rgba(var(--bs-success-rgb), 0.35);
    background: rgba(var(--bs-success-rgb), 0.13);
    color: var(--bs-emphasis-color);
    box-shadow: 0 0.7rem 1.6rem rgba(0, 0, 0, 0.2);
}

.profile-toast i {
    color: rgb(var(--bs-success-rgb));
}

.profile-toast-enter-active,
.profile-toast-leave-active {
    transition: opacity 0.22s ease, transform 0.22s ease;
}

.profile-toast-enter-from,
.profile-toast-leave-to {
    opacity: 0;
    transform: translateY(0.55rem);
}

.profile-loading .placeholder {
    display: block;
    border-radius: 0.5rem;
    min-height: 1rem;
}

@media (max-width: 991.98px) {
    .profile-page {
        max-width: 100%;
    }

    .language-toggle {
        grid-template-columns: 1fr;
    }

    .profile-avatar {
        width: 3rem;
        height: 3rem;
        font-size: 1.05rem;
        border-radius: 0.8rem;
    }

    .profile-save-btn {
        width: 100%;
    }

    .profile-toast-wrapper {
        left: 0.75rem;
        right: 0.75rem;
        bottom: 0.75rem;
    }

    .profile-toast {
        display: flex !important;
        justify-content: center;
        width: 100%;
    }
}
</style>
