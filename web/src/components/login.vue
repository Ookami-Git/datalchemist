<script setup>
import { computed, inject, ref, onMounted, onUnmounted } from 'vue';
import axios from 'axios';
import { useRouter } from 'vue-router';

const router = useRouter();

const username = ref('');
const password = ref('');
const isLoading = ref(false);
const errorMessage = ref('');

const apiUrl = inject('apiUrl');
const parameters = inject('parameters', ref({}));

const appName = computed(() => {
    const raw = parameters.value;
    if (raw && typeof raw === 'object' && !Array.isArray(raw) && raw.name) {
        return raw.name;
    }
    return 'DataChemist';
});

const getSystemTheme = () => {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

const getInitialTheme = () => {
    const fromDoc = document.documentElement.getAttribute('data-bs-theme');
    if (fromDoc === 'light' || fromDoc === 'dark') {
        return fromDoc;
    }
    return getSystemTheme();
};

const themeOverride = ref(getInitialTheme());

const loginTheme = computed(() => {
    return themeOverride.value;
});

const toggleTheme = () => {
    const nextTheme = themeOverride.value === 'dark' ? 'light' : 'dark';
    themeOverride.value = nextTheme;
    document.documentElement.setAttribute('data-bs-theme', nextTheme);
};

let systemThemeMQ = null;
const handleSystemThemeChange = (e) => {
    const nextTheme = e.matches ? 'dark' : 'light';
    themeOverride.value = nextTheme;
    document.documentElement.setAttribute('data-bs-theme', nextTheme);
};

onMounted(() => {
    systemThemeMQ = window.matchMedia('(prefers-color-scheme: dark)');
    systemThemeMQ.addEventListener('change', handleSystemThemeChange);
});

onUnmounted(() => {
    if (systemThemeMQ) {
        systemThemeMQ.removeEventListener('change', handleSystemThemeChange);
    }
});

const loginSceneStyle = computed(() => {
    const raw = parameters.value;
    const hasObject = raw && typeof raw === 'object' && !Array.isArray(raw);
    const theme = loginTheme.value;

    const apiBgA = hasObject ? raw[`bg_color_${theme}`] : '';
    const apiBgB = hasObject ? raw[`bg_color2_${theme}`] : '';

    const fallbackA = theme === 'light' ? '#d6e8f1' : '#123046';
    const fallbackB = theme === 'light' ? '#eff6fb' : '#0a2032';

    return {
        '--login-bg-a': apiBgA || fallbackA,
        '--login-bg-b': apiBgB || fallbackB,
    };
});

const login = async () => {
    if (isLoading.value) {
        return;
    }

    errorMessage.value = '';
    isLoading.value = true;

    try {
        await axios.post(`${apiUrl}/auth/login`, {
            username: username.value.trim(),
            password: password.value
        });

        router.push('/redirect');
    } catch (error) {
        console.error('Authentication failed:', error);
        errorMessage.value = 'Authentication failed. Please check your username and password.';
    } finally {
        isLoading.value = false;
    }
};
</script>

<template>
    <section class="login-page" :style="loginSceneStyle">
        <button class="login-theme-toggle" @click="toggleTheme" :aria-label="themeOverride === 'dark' ? 'Passer au thème clair' : 'Passer au thème sombre'" type="button">
            <i :class="themeOverride === 'dark' ? 'bi bi-sun' : 'bi bi-moon-stars'"></i>
        </button>
        <div class="login-ambient" aria-hidden="true">
            <span class="login-orbit login-orbit-a"><span class="login-orb login-orb-a"></span></span>
            <span class="login-orbit login-orbit-b"><span class="login-orb login-orb-b"></span></span>
            <span class="login-orbit login-orbit-c"><span class="login-orb login-orb-c"></span></span>
            <span class="login-orbit login-orbit-d"><span class="login-orb login-orb-d"></span></span>
            <span class="login-orbit login-orbit-e"><span class="login-orb login-orb-e"></span></span>
            <span class="login-orbit login-orbit-f"><span class="login-orb login-orb-f"></span></span>
            <span class="login-orbit login-orbit-g"><span class="login-orb login-orb-g"></span></span>
            <span class="login-orbit login-orbit-h"><span class="login-orb login-orb-h"></span></span>
        </div>

        <div class="login-container">
            <!-- Côté gauche : Visuel immersif (visible uniquement sur grand écran) -->
            <div class="login-visual d-none d-lg-flex flex-column justify-content-center align-items-center text-center p-5">
                <div class="login-visual-content">
                    <img src="/logo.png" alt="Application logo" class="login-visual-logo mb-4">
                    <h2 class="login-visual-title m-0">{{ appName }}</h2>
                </div>
            </div>

            <!-- Côté droit : Formulaire de connexion -->
            <div class="login-form-container d-flex align-items-center justify-content-center p-4 p-md-5">
                <div class="login-form-box w-100">
                    <div class="text-center d-lg-none mb-4">
                        <img src="/logo.png" alt="Application logo" class="login-logo mb-2">
                        <h2 class="h4 fw-bold">{{ appName }}</h2>
                    </div>

                    <h1 class="h3 fw-bold mb-4">{{ $t('auth.button') }}</h1>

                    <form @submit.prevent="login" novalidate>
                        <div class="form-floating mb-3 login-field">
                            <i class="bi bi-person login-field-icon"></i>
                            <input id="username" v-model="username" type="text" class="form-control"
                                name="username" autocomplete="username" placeholder=" "
                                :disabled="isLoading" required autofocus>
                            <label for="username">{{ $t('global.username') }}</label>
                        </div>

                        <div class="form-floating mb-4 login-field">
                            <i class="bi bi-key login-field-icon"></i>
                            <input id="password" v-model="password" type="password" class="form-control"
                                name="password" autocomplete="current-password" placeholder=" "
                                :disabled="isLoading" required>
                            <label for="password">{{ $t('global.password') }}</label>
                        </div>

                        <div class="d-grid">
                            <button class="btn btn-primary btn-lg btn-login py-3" type="submit" :disabled="isLoading">
                                <span v-if="isLoading" class="spinner-border spinner-border-sm me-2"
                                    role="status" aria-hidden="true"></span>
                                <span class="btn-text">{{ $t('auth.button') }}</span>
                                <i class="bi bi-arrow-right ms-2 btn-icon"></i>
                            </button>
                        </div>

                        <div v-if="errorMessage" class="alert alert-danger mt-4 mb-0" role="alert"
                            aria-live="polite">
                            {{ errorMessage }}
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped lang="scss" src="../scss/login.scss"></style>