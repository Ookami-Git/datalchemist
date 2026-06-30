import test from 'node:test';
import assert from 'node:assert/strict';

import { effectiveGetQuery } from './getVariables.js';

test('effectiveGetQuery uses route query for direct item route', () => {
  assert.deepEqual(
    effectiveGetQuery({}, { getvar: 'getvalue' }),
    { getvar: 'getvalue' }
  );
});

test('effectiveGetQuery keeps explicit preview query first', () => {
  assert.deepEqual(
    effectiveGetQuery({ getvar: 'preview' }, { getvar: 'route' }),
    { getvar: 'preview' }
  );
});
