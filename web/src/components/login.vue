<script setup>
import { computed, inject, ref } from 'vue';
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
    <section class="login-page d-flex align-items-center py-3 py-lg-4">
        <div class="container">
            <div class="row justify-content-center">
                <div class="col-12 col-sm-10 col-md-8 col-lg-6 col-xl-5">
                    <div class="card login-card shadow-sm border-0">
                        <div class="card-body p-4 p-lg-5">
                            <div class="text-center mb-4">
                                <img src="/logo.png" alt="Application logo" class="login-logo">
                                <p class="login-app-name mt-3 mb-2">{{ appName }}</p>
                            </div>

                            <h1 class="h5 fw-semibold text-center mb-2">{{ $t('auth.button') }}</h1>
                            <p class="text-body-secondary text-center mb-4">{{ $t('auth.message') }}</p>

                            <form @submit.prevent="login" novalidate>
                                <div class="mb-3">
                                    <label for="username" class="form-label fw-medium">{{ $t('global.username')
                                        }}</label>
                                    <div class="input-group input-group-lg">
                                        <span class="input-group-text login-input-icon" aria-hidden="true">
                                            <i class="bi bi-person"></i>
                                        </span>
                                        <input id="username" v-model="username" type="text" class="form-control"
                                            name="username" autocomplete="username" :placeholder="$t('global.username')"
                                            :disabled="isLoading" required autofocus>
                                    </div>
                                </div>

                                <div class="mb-4">
                                    <label for="password" class="form-label fw-medium">{{ $t('global.password')
                                        }}</label>
                                    <div class="input-group input-group-lg">
                                        <span class="input-group-text login-input-icon" aria-hidden="true">
                                            <i class="bi bi-key"></i>
                                        </span>
                                        <input id="password" v-model="password" type="password" class="form-control"
                                            name="password" autocomplete="current-password"
                                            :placeholder="$t('global.password')" :disabled="isLoading" required>
                                    </div>
                                </div>

                                <div class="d-grid">
                                    <button class="btn btn-primary btn-lg" type="submit" :disabled="isLoading">
                                        <span v-if="isLoading" class="spinner-border spinner-border-sm me-2"
                                            role="status" aria-hidden="true"></span>
                                        {{ $t('auth.button') }}
                                    </button>
                                </div>

                                <div v-if="errorMessage" class="alert alert-danger mt-3 mb-0" role="alert"
                                    aria-live="polite">
                                    {{ errorMessage }}
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.login-page {
    min-height: calc(100vh - 2rem);
}

.login-card {
    border: 1px solid var(--bs-border-color-translucent);
    background-color: var(--bs-body-bg);
}

.login-logo {
    width: 168px;
    max-width: 62%;
    height: auto;
    object-fit: contain;
}

.login-app-name {
    font-size: 0.77rem;
    letter-spacing: 0.11em;
    text-transform: uppercase;
    color: var(--bs-secondary-color);
    font-weight: 700;
}

.login-input-icon {
    min-width: 2.9rem;
    justify-content: center;
    color: var(--bs-secondary-color);
    background: var(--bs-tertiary-bg);
}

@media (max-width: 575.98px) {
    .login-page {
        min-height: auto;
    }

    .login-card .card-body {
        padding: 1.25rem;
    }

    .login-logo {
        width: 150px;
    }
}
</style>