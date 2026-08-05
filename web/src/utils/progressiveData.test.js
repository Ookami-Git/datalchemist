import assert from 'node:assert/strict';
import test from 'node:test';

import { createDataCollector, mergeSnapshots } from './progressiveData.js';

const plan = {
  get: { page: ['2'] },
  plan: {
    items: {
      10: ['base', 'first'],
      11: ['base'],
      12: []
    }
  }
};

test('un objet sans source est prêt dès le plan', () => {
  const collector = createDataCollector();
  assert.deepEqual(collector.setPlan(plan), ['12']);
  assert.deepEqual(collector.data.get, { page: ['2'] });
});

test('un objet devient prêt dès que ses sources sont arrivées', () => {
  const collector = createDataCollector();
  collector.setPlan(plan);

  // "base" suffit à l'objet 11, pas à l'objet 10 qui attend aussi "first".
  assert.deepEqual(collector.addSource({ name: 'base', id: 1, value: 'value' }).sort(), ['11', '12']);
  assert.deepEqual(collector.addSource({ name: 'first', id: 2, value: 'other' }).sort(), ['10', '11', '12']);

  assert.deepEqual(collector.data.sn, { base: 'value', first: 'other' });
  assert.deepEqual(collector.data.sid, { s1: 'value', s2: 'other' });
});

test('un instantané ne bouge plus après les sources suivantes', () => {
  const collector = createDataCollector();
  collector.setPlan(plan);
  collector.addSource({ name: 'base', id: 1, value: 'value' });

  const snapshot = collector.snapshot();
  collector.addSource({ name: 'first', id: 2, value: 'other' });

  assert.deepEqual(snapshot.sn, { base: 'value' });
});

test('sans plan, aucun objet n\'est déclaré prêt', () => {
  const collector = createDataCollector();
  assert.deepEqual(collector.addSource({ name: 'base', id: 1, value: 'value' }), []);
});

test('mergeSnapshots ne touche pas aux objets déjà servis', () => {
  const first = { sn: { base: 'value' } };
  const second = { sn: { base: 'value', first: 'other' } };

  const current = mergeSnapshots({}, ['10'], first);
  const next = mergeSnapshots(current, ['10', '11'], second);

  assert.equal(next['10'], first, 'un objet affiché garde son instantané');
  assert.equal(next['11'], second);

  // Rien de nouveau : la référence est conservée, aucun rendu n'est relancé.
  assert.equal(mergeSnapshots(next, ['10', '11'], second), next);
});
