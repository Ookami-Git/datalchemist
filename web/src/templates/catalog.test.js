import test from 'node:test';
import assert from 'node:assert/strict';

import { getTemplateDefinition } from './catalog.js';

test('timeline detail field compiles with a valid empty scalar fallback', () => {
  const template = getTemplateDefinition('template/timeline', 1);

  const result = template.compile({
    items: 'events',
    dateField: 'date',
    labelField: 'title',
    detailField: 'detail',
    accent: 'primary'
  });

  assert.equal(result.javascript, '');
  assert.match(result.template, /item\.detail if item is mapping else ''/);
  assert.doesNotMatch(result.template, /else \)/);
});

test('timeline fields keep explicit item jinja expressions', () => {
  const template = getTemplateDefinition('template/timeline', 1);

  const result = template.compile({
    items: 'events',
    dateField: '{{ item.date | date("DD/MM/YYYY") }}',
    labelField: '{{ item.title }}',
    detailField: '{{ item.detail }}',
    accent: 'primary'
  });

  assert.match(result.template, /\{\{ item\.date \| date\("DD\/MM\/YYYY"\) \}\}/);
  assert.match(result.template, /\{\{ item\.title \}\}/);
  assert.match(result.template, /\{\{ item\.detail \}\}/);
  assert.doesNotMatch(result.template, /item\.\{\{/);
});

test('status color rule can compare an explicit item jinja expression', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const result = template.compile({
    items: 'services',
    labelField: '{{ item.name }}',
    badgeField: '{{ item.status }}',
    colorRules: [
      { field: '{{ item.status }}', operator: 'eq', value: 'OK', accent: 'success' }
    ]
  });

  assert.match(result.template, /\(item\.status \| default\(''\)\) == 'OK'/);
  assert.match(result.template, /\{\{ item\.name \}\}/);
  assert.match(result.template, /\{\{ item\.status \}\}/);
});
