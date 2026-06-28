import test from 'node:test';
import assert from 'node:assert/strict';
import nunjucks from 'nunjucks';

import {
  compileTemplateDefinition,
  getTemplateDefinition,
  migrateTemplateConfig,
  validateTemplateConfig
} from './catalog.js';

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

test('template validation reports missing required fields', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const issues = validateTemplateConfig(template, {
    title: 'Services',
    items: 'services'
  });

  assert.deepEqual(issues.map((issue) => issue.field), ['labelField', 'badgeField']);
});

test('catalog facade normalizes config before compilation', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const normalized = migrateTemplateConfig(template, {
    title: 'Clients',
    items: 'clients',
    pageLength: 'not-a-number',
    searching: 'TRUE',
    buttons: 'missing'
  }, 1);
  const compiled = compileTemplateDefinition(template, normalized);

  assert.equal(normalized.pageLength, 25);
  assert.equal(normalized.searching, 'true');
  assert.equal(normalized.buttons, 'basic');
  assert.match(compiled.template, /data-dc-datatable=/);
});

test('compact table keeps DataTables initialization in javascript output', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    buttons: 'all',
    pageLength: 10
  });

  assert.match(compiled.template, /data-dc-datatable=/);
  assert.doesNotMatch(compiled.template, /new DataTable/);
  assert.doesNotMatch(compiled.template, /DataTable\.Buttons/);
  assert.match(compiled.javascript, /const dataTableOptions = /);
  assert.match(compiled.javascript, /"layout":\{"topStart":\["buttons"\],"topEnd":\["pageLength","search"\],"bottomStart":"info","bottomEnd":"paging"\}/);
  assert.doesNotMatch(compiled.javascript, /"dom":/);
  assert.match(compiled.javascript, /const tableOptions = \{ \.\.\.dataTableOptions \}/);
  assert.match(compiled.javascript, /new DataTable\(table, tableOptions\)/);
  assert.match(compiled.javascript, /DataTable\.Buttons\.jszip/);
});

test('compact table does not require a display title', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const issues = validateTemplateConfig(template, {
    items: 'clients'
  });
  const compiled = compileTemplateDefinition(template, {
    items: 'clients'
  }, {
    item: { id: 12, name: 'Clients' }
  });

  assert.deepEqual(issues, []);
  assert.match(compiled.template, /data-dc-datatable=/);
});

test('compact table disables dependent DataTables options with their parent option', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    searching: 'false',
    searchBuilder: 'true',
    searchPanes: 'true',
    paging: 'false',
    info: 'true',
    ordering: 'false',
    orderColumn: 2,
    scrollX: 'false',
    fixedColumnsLeft: 1
  });

  assert.match(compiled.javascript, /searching":false/);
  assert.doesNotMatch(compiled.javascript, /searchBuilder":/);
  assert.doesNotMatch(compiled.javascript, /searchPanes":/);
  assert.match(compiled.javascript, /paging":false/);
  assert.match(compiled.javascript, /info":false/);
  assert.doesNotMatch(compiled.javascript, /fixedColumns":/);
  assert.match(compiled.javascript, /order":\[\]/);
});

test('compact table uses DataTables layout feature names for advanced controls', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    searchBuilder: 'true',
    searchPanes: 'true',
    buttons: 'all'
  });

  assert.match(compiled.javascript, /"topStart":\["searchBuilder","searchPanes","buttons"\]/);
  assert.match(compiled.javascript, /"searchPanes":\{"threshold":0\.6\}/);
  assert.doesNotMatch(compiled.javascript, /"topStart":\["Q","P","buttons"\]/);
});

test('compact table supports manual SearchPanes columns', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    searchPanes: 'true',
    searchPanesAutomatic: 'false',
    searchPanesColumns: JSON.stringify(['Segment', 'Statut'])
  });

  assert.match(compiled.javascript, /"searchPanes":\{"columns":\[\],"threshold":1\}/);
  assert.match(compiled.javascript, /const dataTableSearchPanesColumns = \["Segment","Statut"\]/);
  assert.match(compiled.javascript, /searchPanes: \{ show: true, threshold: 1 \}/);
});

test('compact table clamps automatic SearchPanes threshold', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    searchPanes: 'true',
    searchPanesThreshold: 2
  });

  assert.match(compiled.javascript, /"searchPanes":\{"threshold":1\}/);
});

test('compact table resolves initial sort column from configured column key', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    columns: JSON.stringify([
      { key: 'name', label: 'Client' },
      { key: 'revenue', label: 'CA' },
      { key: 'segment', label: 'Segment' }
    ]),
    orderColumn: 'revenue',
    orderDirection: 'desc'
  });

  assert.match(compiled.javascript, /"order":\[\[1,"desc"\]\]/);
});

test('compact table keeps legacy numeric initial sort column values', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    columns: JSON.stringify([
      { key: 'name', label: 'Client' },
      { key: 'revenue', label: 'CA' },
      { key: 'segment', label: 'Segment' }
    ]),
    orderColumn: 3
  });

  assert.match(compiled.javascript, /"order":\[\[2,"asc"\]\]/);
});

test('compact table can resolve automatic initial sort column at runtime', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    orderColumn: 'revenue'
  });

  assert.match(compiled.javascript, /const dataTableOrderColumn = "revenue"/);
  assert.match(compiled.javascript, /querySelectorAll\('thead th'\)/);
  assert.match(compiled.javascript, /tableOptions\.order = \[\[headerIndex, dataTableOrderDirection\]\]/);
});

test('compact table enables DataTables runtime controls consistently', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    pageLength: 50,
    ordering: 'true',
    orderColumn: 2,
    orderDirection: 'desc',
    scrollX: 'true',
    selectRows: 'true'
  });

  assert.match(compiled.javascript, /"pageLength":50/);
  assert.match(compiled.javascript, /"lengthMenu":\[\[10,25,50,100,-1\],\[10,25,50,100,"All"\]\]/);
  assert.match(compiled.javascript, /"lengthChange":true/);
  assert.match(compiled.javascript, /"ordering":true/);
  assert.match(compiled.javascript, /"order":\[\[1,"desc"\]\]/);
  assert.match(compiled.javascript, /"select":\{"style":"multi","items":"row"\}/);
  assert.match(compiled.javascript, /"scrollX":true/);
  assert.match(compiled.javascript, /delete data\.length/);
  assert.match(compiled.javascript, /delete data\.start/);
  assert.doesNotMatch(compiled.javascript, /"responsive":true/);
  assert.match(compiled.template, /text-nowrap/);
});

test('compact table treats Scroller as an alternative to visible pagination', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    paging: 'false',
    scrollX: 'false',
    scroller: 'true'
  });

  assert.match(compiled.javascript, /"paging":true/);
  assert.match(compiled.javascript, /"lengthChange":false/);
  assert.match(compiled.javascript, /"bottomEnd":null/);
  assert.match(compiled.javascript, /"scroller":true/);
  assert.match(compiled.javascript, /"scrollY":"50vh"/);
  assert.match(compiled.javascript, /"scrollCollapse":true/);
  assert.doesNotMatch(compiled.javascript, /"scrollX":true/);
  assert.match(compiled.javascript, /scrollBody\.style\.overflowX = 'hidden'/);
  assert.doesNotMatch(compiled.template, /text-nowrap/);
});

test('compact table keeps pagination ahead of Scroller for conflicting saved configs', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    paging: 'true',
    scroller: 'true'
  });

  assert.match(compiled.javascript, /"paging":true/);
  assert.match(compiled.javascript, /"bottomEnd":"paging"/);
  assert.doesNotMatch(compiled.javascript, /"scroller":true/);
  assert.doesNotMatch(compiled.javascript, /"scrollY":"50vh"/);
});

test('compact table includes custom page length in the DataTables length menu', () => {
  const template = getTemplateDefinition('template/compact-table', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'clients',
    pageLength: 30
  });

  assert.match(compiled.javascript, /"pageLength":30/);
  assert.match(compiled.javascript, /"lengthMenu":\[\[10,25,30,50,100,-1\],\[10,25,30,50,100,"All"\]\]/);
});

test('catalog facade formats compiled visual templates for free mode readability', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const compiled = compileTemplateDefinition(template, {
    items: 'services',
    labelField: 'name',
    detailField: 'detail',
    badgeField: 'status'
  });

  assert.match(compiled.template, /\n      <li class="list-group-item/);
  assert.match(compiled.template, /\n        <div>/);
  assert.match(compiled.template, /\n    \{% endfor %\}/);

  const rendered = nunjucks.renderString(compiled.template, {
    services: [{ name: 'API', detail: 'Latence 42 ms', status: 'OK' }]
  });
  assert.match(rendered, /API/);
  assert.match(rendered, /Latence 42 ms/);
  assert.match(rendered, /OK/);
});

test('visual templates compile global variables before the template body', () => {
  const template = getTemplateDefinition('template/progress-card', 1);

  const result = template.compile({
    title: 'Objectif',
    current: '{{ var.global.done }}',
    target: '{{ var.global.goal }}',
    percent: '{{ var.global.ratio }}',
    globalVariables: JSON.stringify([
      { name: 'done', value: '{{ stats.done }}' },
      { name: 'goal', value: '{{ stats.goal }}' },
      { name: 'ratio', value: '{{ (var.global.done | float / var.global.goal | float) * 100 }}' }
    ])
  });

  assert.match(result.template, /^\{% set _dcGlobal_done = stats\.done %\}\{% set _dcGlobalVar = \{"done": _dcGlobal_done\} %\}\{% set _dcLoopVar = \{\} %\}\{% set var = \{"global": _dcGlobalVar, "loop": _dcLoopVar\} %\}/);
  assert.match(result.template, /\{% set _dcGlobal_ratio = \(var\.global\.done \| float \/ var\.global\.goal \| float\) \* 100 %\}/);
  assert.match(result.template, /\{\{ var\.global\.ratio \}\}%/);

  const rendered = nunjucks.renderString(result.template, { stats: { done: 75, goal: 100 } });
  assert.match(rendered, /75 \/ 100/);
  assert.match(rendered, /75%/);
});

test('collection templates compile loop variables inside each iteration', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const result = template.compile({
    items: 'services',
    labelField: '{{ var.loop.label }}',
    badgeField: '{{ var.loop.badge }}',
    globalVariables: [{ name: 'fallbackBadge', value: 'UNKNOWN' }],
    loopVariables: [
      { name: 'label', value: "{{ item.name if item is mapping else key }}" },
      { name: 'badge', value: "{{ item.status | default(var.global.fallbackBadge) }}" }
    ]
  });

  assert.match(result.template, /^\{% set _dcGlobal_fallbackBadge %\}UNKNOWN\{% endset %\}\{% set _dcGlobalVar = \{"fallbackBadge": _dcGlobal_fallbackBadge\} %\}/);
  assert.match(result.template, /\{% set _dcRowValue = item %\}\{% set _dcGlobalVar = \{"fallbackBadge": _dcGlobal_fallbackBadge\} %\}\{% set _dcLoopVar = \{\} %\}/);
  assert.match(result.template, /\{% set _dcLoop_label = item\.name if item is mapping else key %\}\{% set _dcGlobalVar = \{"fallbackBadge": _dcGlobal_fallbackBadge\} %\}\{% set _dcLoopVar = \{"label": _dcLoop_label\} %\}/);
  assert.match(result.template, /\{\{ var\.loop\.label \}\}/);
  assert.match(result.template, /\{\{ var\.loop\.badge \}\}/);

  const rendered = nunjucks.renderString(result.template, { services: [{ name: 'API', status: 'OK' }, { name: 'Jobs' }] });
  assert.match(rendered, /API/);
  assert.match(rendered, /OK/);
  assert.match(rendered, /UNKNOWN/);
});

test('invalid template variable names are ignored', () => {
  const template = getTemplateDefinition('template/timeline', 1);

  const result = template.compile({
    items: 'events',
    dateField: 'date',
    labelField: '{{ var.global.clean }}',
    globalVariables: [
      { name: 'bad-name', value: '{{ value }}' },
      { name: 'clean', value: '{{ title }}' }
    ]
  });

  assert.doesNotMatch(result.template, /bad-name/);
  assert.match(result.template, /^\{% set _dcGlobal_clean = title %\}\{% set _dcGlobalVar = \{"clean": _dcGlobal_clean\} %\}/);
});

test('template variables preserve expression values for conditions', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const result = template.compile({
    items: 'services',
    labelField: '{{ var.loop.label }}',
    badgeField: '{% if var.loop.score >= 10 %}OK{% else %}KO{% endif %}',
    loopVariables: [
      { name: 'label', value: '{{ item.name }}' },
      { name: 'score', value: '{{ item.score }}' }
    ]
  });

  const rendered = nunjucks.renderString(result.template, {
    services: [
      { name: 'API', score: 12 },
      { name: 'Jobs', score: 4 }
    ]
  });

  assert.match(rendered, /API/);
  assert.match(rendered, /OK/);
  assert.match(rendered, /Jobs/);
  assert.match(rendered, /KO/);
});

test('loop variables do not overwrite global variables with the same name', () => {
  const template = getTemplateDefinition('template/status-list', 1);

  const result = template.compile({
    items: 'services',
    labelField: '{{ var.global.status }}:{{ var.loop.status }}',
    badgeField: '{{ var.loop.status }}',
    globalVariables: [
      { name: 'status', value: 'GLOBAL' }
    ],
    loopVariables: [
      { name: 'status', value: '{{ item.status }}' }
    ]
  });

  const rendered = nunjucks.renderString(result.template, {
    services: [{ status: 'LOOP' }]
  });

  assert.match(rendered, /GLOBAL:LOOP/);
});
