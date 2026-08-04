import test from 'node:test';
import assert from 'node:assert/strict';

import { isPasswordValid, passwordRuleStates, passwordStrength, PASSWORD_MIN_LENGTH } from './password.js';

// These cases mirror TestValidatePassword in database/database_test.go: the two
// implementations must agree, otherwise the form accepts what the server refuses.
test('isPasswordValid rejects passwords that break a rule', () => {
  for (const password of ['', 'Short-1a', 'UPPERCASE-123', 'lowercase-123', 'NoDigitHere-x', 'NoSpecialHere1']) {
    assert.equal(isPasswordValid(password), false, `${password} should be rejected`);
  }
});

test('isPasswordValid accepts compliant passwords', () => {
  for (const password of ['Initial-Password1', 'Sup3r Sécurisé!']) {
    assert.equal(isPasswordValid(password), true, `${password} should be accepted`);
  }
});

test('passwordRuleStates reports which rule fails', () => {
  const states = passwordRuleStates('nodigits-or-upper');
  const failing = states.filter((rule) => !rule.valid).map((rule) => rule.id);
  assert.deepEqual(failing, ['upper', 'digit']);
});

test('the length rule follows PASSWORD_MIN_LENGTH', () => {
  const shortest = `Aa1-${'x'.repeat(PASSWORD_MIN_LENGTH - 4)}`;
  assert.equal(isPasswordValid(shortest), true);
  assert.equal(isPasswordValid(shortest.slice(0, -1)), false);
});

test('passwordStrength grows with quality', () => {
  assert.equal(passwordStrength('').level, 0);
  assert.equal(passwordStrength('abc').level, 1);
  assert.equal(passwordStrength('Initial-Pwd1').level, 3);
  assert.equal(passwordStrength('Initial-Password1').level, 4);
});
