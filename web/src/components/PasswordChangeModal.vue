<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, inject } from 'vue';
import { Modal } from 'bootstrap';
import axios from 'axios';
import PasswordPolicy from '@/components/PasswordPolicy.vue';
import { isPasswordValid } from '@/utils/password';

const apiUrl = inject('apiUrl');

const props = defineProps({
    // Only used to give password managers the account being edited.
    username: {
        type: String,
        default: '',
    },
});

const emit = defineEmits(['changed']);

const modalElement = ref(null);
const currentPasswordInput = ref(null);
let modalInstance = null;

const currentPassword = ref('');
const newPassword = ref('');
const confirmation = ref('');
const confirmationTouched = ref(false);
const visibleFields = ref({ current: false, next: false, confirm: false });
const isSubmitting = ref(false);
const errorCode = ref('');
const errorMessage = ref('');

// Server error codes we can phrase better than the raw API message.
const KNOWN_ERROR_CODES = [
    'invalid_current_password',
    'password_unchanged',
    'weak_password',
    'not_local',
];

const newPasswordIsValid = computed(() => isPasswordValid(newPassword.value));
const confirmationMatches = computed(() => !!confirmation.value && confirmation.value === newPassword.value);

// Stay quiet while the confirmation can still become a match, so the field does
// not turn red on every keystroke, but never leave the submit button disabled
// without telling the user why.
const confirmationMismatch = computed(() => {
    if (!confirmation.value || confirmationMatches.value) {
        return false;
    }
    return confirmationTouched.value || !newPassword.value.startsWith(confirmation.value);
});
const isSameAsCurrent = computed(() => !!newPassword.value && newPassword.value === currentPassword.value);

const canSubmit = computed(() =>
    !isSubmitting.value
    && !!currentPassword.value
    && newPasswordIsValid.value
    && confirmationMatches.value
    && !isSameAsCurrent.value
);

const currentPasswordIsWrong = computed(() => errorCode.value === 'invalid_current_password');

const feedback = computed(() => {
    if (!errorCode.value && !errorMessage.value) {
        return null;
    }
    if (KNOWN_ERROR_CODES.includes(errorCode.value)) {
        return { key: `password.error.${errorCode.value}`, text: '' };
    }
    return { key: 'password.error.unknown', text: errorMessage.value };
});

function toggleField(field) {
    visibleFields.value[field] = !visibleFields.value[field];
}

function resetForm() {
    currentPassword.value = '';
    newPassword.value = '';
    confirmation.value = '';
    confirmationTouched.value = false;
    visibleFields.value = { current: false, next: false, confirm: false };
    errorCode.value = '';
    errorMessage.value = '';
    isSubmitting.value = false;
}

function open() {
    resetForm();
    modalInstance?.show();
}

async function submit() {
    if (!canSubmit.value) {
        return;
    }

    isSubmitting.value = true;
    errorCode.value = '';
    errorMessage.value = '';

    try {
        await axios.post(`${apiUrl}/user/password`, {
            current_password: currentPassword.value,
            new_password: newPassword.value,
        });
    } catch (error) {
        errorCode.value = error?.response?.data?.code || '';
        errorMessage.value = error?.response?.data?.error || '';
        isSubmitting.value = false;

        // Wrong current password: keep everything else, just send the user back
        // to the single field they have to fix.
        if (currentPasswordIsWrong.value) {
            currentPassword.value = '';
            await nextTick();
            currentPasswordInput.value?.focus();
        }
        return;
    }

    isSubmitting.value = false;
    modalInstance?.hide();
    emit('changed');
}

function handleShown() {
    currentPasswordInput.value?.focus();
}

function handleHidden() {
    resetForm();
}

onMounted(() => {
    modalInstance = new Modal(modalElement.value);
    modalElement.value.addEventListener('shown.bs.modal', handleShown);
    modalElement.value.addEventListener('hidden.bs.modal', handleHidden);
});

onBeforeUnmount(() => {
    modalElement.value?.removeEventListener('shown.bs.modal', handleShown);
    modalElement.value?.removeEventListener('hidden.bs.modal', handleHidden);
    modalInstance?.dispose();
    modalInstance = null;
});

defineExpose({ open });
</script>

<template>
    <div ref="modalElement" class="modal fade password-modal" tabindex="-1" aria-labelledby="passwordModalTitle"
        aria-hidden="true">
        <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
                <form @submit.prevent="submit" novalidate>
                    <div class="modal-header">
                        <div class="d-flex align-items-center gap-3">
                            <span class="password-modal-icon" aria-hidden="true"><i class="bi bi-shield-lock"></i></span>
                            <div>
                                <h1 class="modal-title fs-5 mb-0" id="passwordModalTitle">{{ $t('password.title') }}</h1>
                                <p class="mb-0 small text-secondary">{{ $t('password.subtitle') }}</p>
                            </div>
                        </div>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" :aria-label="$t('global.close')"
                            :disabled="isSubmitting"></button>
                    </div>

                    <div class="modal-body">
                        <!-- Hidden field so password managers know which account is edited. -->
                        <input type="text" :value="props.username" autocomplete="username" class="visually-hidden"
                            tabindex="-1" readonly>

                        <div class="mb-4">
                            <label for="passwordCurrent" class="form-label">{{ $t('password.current') }}</label>
                            <div class="input-group">
                                <span class="input-group-text"><i class="bi bi-key"></i></span>
                                <input ref="currentPasswordInput" id="passwordCurrent" v-model="currentPassword"
                                    :type="visibleFields.current ? 'text' : 'password'" class="form-control"
                                    :class="{ 'is-invalid': currentPasswordIsWrong }" autocomplete="current-password"
                                    :disabled="isSubmitting" required>
                                <button type="button" class="btn btn-outline-secondary" @click="toggleField('current')"
                                    :aria-label="visibleFields.current ? $t('password.hide') : $t('password.show')"
                                    :disabled="isSubmitting">
                                    <i :class="visibleFields.current ? 'bi bi-eye-slash' : 'bi bi-eye'"></i>
                                </button>
                            </div>
                            <p v-if="currentPasswordIsWrong" class="form-text text-danger mb-0">
                                {{ $t('password.error.invalid_current_password') }}
                            </p>
                        </div>

                        <hr class="my-4">

                        <div class="mb-3">
                            <label for="passwordNew" class="form-label">{{ $t('password.new') }}</label>
                            <div class="input-group">
                                <span class="input-group-text"><i class="bi bi-lock"></i></span>
                                <input id="passwordNew" v-model="newPassword"
                                    :type="visibleFields.next ? 'text' : 'password'" class="form-control"
                                    :class="{ 'is-invalid': isSameAsCurrent }" autocomplete="new-password"
                                    :disabled="isSubmitting" required>
                                <button type="button" class="btn btn-outline-secondary" @click="toggleField('next')"
                                    :aria-label="visibleFields.next ? $t('password.hide') : $t('password.show')"
                                    :disabled="isSubmitting">
                                    <i :class="visibleFields.next ? 'bi bi-eye-slash' : 'bi bi-eye'"></i>
                                </button>
                            </div>
                            <p v-if="isSameAsCurrent" class="form-text text-danger mb-0">
                                {{ $t('password.error.password_unchanged') }}
                            </p>
                        </div>

                        <PasswordPolicy :password="newPassword" class="mb-4" />

                        <div>
                            <label for="passwordConfirm" class="form-label">{{ $t('password.confirm') }}</label>
                            <div class="input-group">
                                <span class="input-group-text"><i class="bi bi-lock-fill"></i></span>
                                <input id="passwordConfirm" v-model="confirmation"
                                    :type="visibleFields.confirm ? 'text' : 'password'" class="form-control"
                                    :class="{ 'is-invalid': confirmationMismatch }" autocomplete="new-password"
                                    :disabled="isSubmitting" required @blur="confirmationTouched = true">
                                <button type="button" class="btn btn-outline-secondary" @click="toggleField('confirm')"
                                    :aria-label="visibleFields.confirm ? $t('password.hide') : $t('password.show')"
                                    :disabled="isSubmitting">
                                    <i :class="visibleFields.confirm ? 'bi bi-eye-slash' : 'bi bi-eye'"></i>
                                </button>
                            </div>
                            <p v-if="confirmationMismatch" class="form-text text-danger mb-0">
                                {{ $t('password.mismatch') }}
                            </p>
                            <p v-else-if="confirmationMatches" class="form-text text-success mb-0">
                                <i class="bi bi-check2 me-1"></i>{{ $t('password.match') }}
                            </p>
                        </div>

                        <div v-if="feedback && !currentPasswordIsWrong" class="alert alert-danger mt-4 mb-0"
                            role="alert" aria-live="assertive">
                            <i class="bi bi-exclamation-triangle-fill me-2"></i>
                            <span>{{ $t(feedback.key) }}</span>
                            <span v-if="feedback.text" class="d-block small mt-1">{{ feedback.text }}</span>
                        </div>
                    </div>

                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal"
                            :disabled="isSubmitting">
                            {{ $t('global.cancel') }}
                        </button>
                        <button type="submit" class="btn btn-primary" :disabled="!canSubmit">
                            <span v-if="isSubmitting" class="spinner-border spinner-border-sm me-2" role="status"
                                aria-hidden="true"></span>
                            <i v-else class="bi bi-shield-check me-2"></i>
                            {{ $t('password.submit') }}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<style scoped>
.password-modal .modal-content {
    border: 1px solid var(--bs-border-color-translucent);
    border-radius: 1rem;
}

.password-modal .modal-header {
    align-items: flex-start;
}

.password-modal-icon {
    display: grid;
    place-items: center;
    width: 2.6rem;
    height: 2.6rem;
    border-radius: 0.8rem;
    font-size: 1.15rem;
    color: rgb(var(--bs-primary-rgb));
    background: rgba(var(--bs-primary-rgb), 0.12);
    flex-shrink: 0;
}

.password-modal .input-group-text {
    background-color: var(--bs-tertiary-bg);
    border-color: var(--bs-border-color);
    color: var(--bs-secondary-color);
}

.password-modal .form-control:focus {
    border-color: rgba(var(--bs-primary-rgb), 0.45);
    box-shadow: 0 0 0 0.2rem rgba(var(--bs-primary-rgb), 0.15);
}

.password-modal .form-label {
    font-weight: 600;
    margin-bottom: 0.45rem;
}
</style>
