<script setup>
import { computed } from 'vue';
import { passwordRuleStates, passwordStrength } from '@/utils/password';

const props = defineProps({
    password: {
        type: String,
        default: '',
    },
    showStrength: {
        type: Boolean,
        default: true,
    },
});

const rules = computed(() => passwordRuleStates(props.password));
const strength = computed(() => passwordStrength(props.password));
const isEmpty = computed(() => !props.password);
</script>

<template>
    <div class="password-policy">
        <div v-if="showStrength" class="password-strength" :class="`level-${strength.level}`">
            <div class="password-strength-track" role="presentation">
                <span v-for="step in 4" :key="step" class="password-strength-step"
                    :class="{ 'is-filled': step <= strength.level }"></span>
            </div>
            <span class="password-strength-label">
                {{ strength.labelKey ? $t(strength.labelKey) : $t('password.strength.empty') }}
            </span>
        </div>

        <ul class="password-rules list-unstyled mb-0" aria-live="polite">
            <li v-for="rule in rules" :key="rule.id" class="password-rule"
                :class="{ 'is-valid': rule.valid, 'is-pending': isEmpty }">
                <i :class="rule.valid ? 'bi bi-check-circle-fill' : 'bi bi-circle'" aria-hidden="true"></i>
                <span>{{ $t(rule.labelKey, rule.params) }}</span>
            </li>
        </ul>
    </div>
</template>

<style scoped>
.password-strength {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    margin-bottom: 0.75rem;
}

.password-strength-track {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.25rem;
    flex-grow: 1;
}

.password-strength-step {
    height: 0.3rem;
    border-radius: 0.3rem;
    background: var(--bs-tertiary-bg);
    border: 1px solid var(--bs-border-color-translucent);
    transition: background-color 0.25s ease, border-color 0.25s ease;
}

.password-strength-label {
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--bs-secondary-color);
    min-width: 5.5rem;
    text-align: right;
}

.password-strength.level-1 .password-strength-step.is-filled {
    background: var(--bs-danger);
    border-color: var(--bs-danger);
}

.password-strength.level-2 .password-strength-step.is-filled {
    background: var(--bs-warning);
    border-color: var(--bs-warning);
}

.password-strength.level-3 .password-strength-step.is-filled,
.password-strength.level-4 .password-strength-step.is-filled {
    background: var(--bs-success);
    border-color: var(--bs-success);
}

.password-strength.level-1 .password-strength-label {
    color: var(--bs-danger);
}

.password-strength.level-2 .password-strength-label {
    color: var(--bs-warning-text-emphasis);
}

.password-strength.level-3 .password-strength-label,
.password-strength.level-4 .password-strength-label {
    color: var(--bs-success-text-emphasis);
}

.password-rules {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.25rem 0.9rem;
}

.password-rule {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.82rem;
    color: var(--bs-danger);
    transition: color 0.2s ease;
}

.password-rule i {
    font-size: 0.72rem;
    flex-shrink: 0;
}

.password-rule.is-pending {
    color: var(--bs-secondary-color);
}

.password-rule.is-valid {
    color: var(--bs-success-text-emphasis);
}

@media (max-width: 575.98px) {
    .password-rules {
        grid-template-columns: 1fr;
    }

    .password-strength-label {
        min-width: 4.5rem;
    }
}
</style>
