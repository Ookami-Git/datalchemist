// Password policy shared by every screen where a user picks a password
// (first-run setup, profile). It mirrors database.ValidatePassword in
// database/database.go: keep both in sync, the server always has the last word.

export const PASSWORD_MIN_LENGTH = 12;

// The `special` rule is the complement of the three others, exactly like the
// server which classifies each rune as lower / upper / digit / anything else.
export const PASSWORD_RULES = [
    {
        id: 'length',
        labelKey: 'password.rule.length',
        params: { length: PASSWORD_MIN_LENGTH },
        test: (value) => [...value].length >= PASSWORD_MIN_LENGTH,
    },
    {
        id: 'lower',
        labelKey: 'password.rule.lower',
        test: (value) => /\p{Ll}/u.test(value),
    },
    {
        id: 'upper',
        labelKey: 'password.rule.upper',
        test: (value) => /\p{Lu}/u.test(value),
    },
    {
        id: 'digit',
        labelKey: 'password.rule.digit',
        test: (value) => /\p{Nd}/u.test(value),
    },
    {
        id: 'special',
        labelKey: 'password.rule.special',
        test: (value) => /[^\p{Ll}\p{Lu}\p{Nd}]/u.test(value),
    },
];

export function passwordRuleStates(password) {
    const value = password || '';
    return PASSWORD_RULES.map((rule) => ({
        id: rule.id,
        labelKey: rule.labelKey,
        params: rule.params || {},
        valid: rule.test(value),
    }));
}

export function isPasswordValid(password) {
    return passwordRuleStates(password).every((rule) => rule.valid);
}

const STRENGTH_LABELS = [
    '',
    'password.strength.weak',
    'password.strength.fair',
    'password.strength.good',
    'password.strength.strong',
];

// Indicative meter only: the rules above are what actually gates the form.
export function passwordStrength(password) {
    const value = password || '';
    if (!value) {
        return { level: 0, labelKey: '' };
    }

    const satisfied = passwordRuleStates(value).filter((rule) => rule.valid).length;

    let level = 1;
    if (satisfied >= 3) {
        level = 2;
    }
    if (satisfied === PASSWORD_RULES.length) {
        level = [...value].length >= 16 ? 4 : 3;
    }

    return { level, labelKey: STRENGTH_LABELS[level] };
}
