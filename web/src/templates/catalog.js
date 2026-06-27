function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

const accentOptions = [
  { value: 'primary', label: 'Bleu' },
  { value: 'success', label: 'Vert' },
  { value: 'info', label: 'Cyan' },
  { value: 'warning', label: 'Jaune' },
  { value: 'danger', label: 'Rouge' },
  { value: 'secondary', label: 'Gris' }
];
const accentValues = accentOptions.map((option) => option.value);
const htmlColorPattern = /^#[0-9a-f]{6}$/i;
const booleanOptions = [
  { value: 'true', label: 'Oui' },
  { value: 'false', label: 'Non' }
];
const dataTableButtonOptions = [
  { value: 'none', label: 'Aucun' },
  { value: 'basic', label: 'Export simple' },
  { value: 'all', label: 'Export complet' }
];
const orderDirectionOptions = [
  { value: 'asc', label: 'Ascendant' },
  { value: 'desc', label: 'Descendant' }
];

function normalizeAccent(value, fallback = 'primary') {
  if (value && typeof value === 'object' && value.baseField !== undefined) {
    return value;
  }
  const str = String(value ?? '').trim();
  if (str.includes('{') || str.includes('%')) {
    return str;
  }
  if (htmlColorPattern.test(str)) {
    return str;
  }
  return accentValues.includes(str) ? str : fallback;
}

function accentCssColor(value) {
  const normalized = String(value ?? '').trim();
  if (normalized.includes('{') || normalized.includes('%')) {
    return normalized;
  }
  if (htmlColorPattern.test(normalized)) {
    return normalized;
  }
  return accentValues.includes(normalized) ? `var(--bs-${normalized})` : `var(--bs-primary)`;
}

function normalizeProgressColorRules(value) {
  if (!Array.isArray(value)) return [];
  return value
    .map((rule) => ({
      operator: ['lt', 'lte', 'gt', 'gte', 'between'].includes(rule?.operator) ? rule.operator : 'lt',
      value: String(rule?.value ?? '').trim(),
      valueMax: String(rule?.valueMax ?? '').trim(),
      accent: normalizeAccent(rule?.accent, 'info')
    }))
    .filter((rule) => rule.value !== '' && (rule.operator !== 'between' || rule.valueMax !== ''));
}

function compileProgressAccentRules(rules, fallbackAccent) {
  if (!rules.length) return fallbackAccent;
  const conditions = rules.map((rule) => {
    const min = Number.parseFloat(rule.value);
    const max = Number.parseFloat(rule.valueMax);
    if (!Number.isFinite(min)) return null;
    if (rule.operator === 'between') {
      if (!Number.isFinite(max)) return null;
      return `progressPercent >= ${min} and progressPercent <= ${max}`;
    }
    if (rule.operator === 'lt') return `progressPercent < ${min}`;
    if (rule.operator === 'lte') return `progressPercent <= ${min}`;
    if (rule.operator === 'gt') return `progressPercent > ${min}`;
    if (rule.operator === 'gte') return `progressPercent >= ${min}`;
    return null;
  });
  const branches = rules
    .map((rule, index) => ({ rule, condition: conditions[index] }))
    .filter(({ condition }) => condition)
    .map(({ rule, condition }, index) => `${index === 0 ? 'if' : 'elif'} ${condition} %}${accentCssColor(rule.accent)}{% `)
    .join('');

  return branches
    ? `{% ${branches}else %}${accentCssColor(fallbackAccent)}{% endif %}`
    : accentCssColor(fallbackAccent);
}

function quoteNunjucksString(value) {
  return `'${String(value ?? '').replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`;
}

function statusRuleValueTemplate(value) {
  const normalized = String(value ?? '').trim();
  if (normalized.includes('{{') || normalized.includes('{%')) {
    return normalized;
  }
  return escapeHtml(normalized);
}

function normalizeStatusColorRules(value) {
  if (!Array.isArray(value)) return [];
  return value
    .map((rule) => ({
      field: dataPath(rule?.field, ''),
      operator: ['eq', 'neq', 'contains', 'not_contains', 'empty', 'not_empty', 'lt', 'lte', 'gt', 'gte', 'between'].includes(rule?.operator)
        ? rule.operator
        : 'contains',
      value: String(rule?.value ?? '').trim(),
      valueMax: String(rule?.valueMax ?? '').trim(),
      accent: normalizeAccent(rule?.accent, 'success')
    }))
    .filter((rule) => ['empty', 'not_empty'].includes(rule.operator) || (rule.value !== '' && (rule.operator !== 'between' || rule.valueMax !== '')));
}

function compileStatusAccentRules(rules, fallbackAccent, fallbackField) {
  if (!rules.length) return fallbackAccent;
  const conditions = rules.map((rule) => {
    const expectedVar = `statusExpected${rules.indexOf(rule)}`;
    const expectedMaxVar = `statusExpectedMax${rules.indexOf(rule)}`;
    const fieldExpression = itemFieldExpression(rule.field, fallbackField);
    const textExpression = `(${fieldExpression} | default(''))`;
    const numberExpression = `(${fieldExpression} | default(0) | float)`;
    const expected = rule.value.includes('{{') || rule.value.includes('{%') ? expectedVar : quoteNunjucksString(rule.value);
    const expectedNumber = rule.value.includes('{{') || rule.value.includes('{%')
      ? `(${expectedVar} | default(0) | float)`
      : Number.parseFloat(rule.value);
    const expectedMaxNumber = rule.valueMax.includes('{{') || rule.valueMax.includes('{%')
      ? `(${expectedMaxVar} | default(0) | float)`
      : Number.parseFloat(rule.valueMax);
    if (rule.operator === 'eq') return `${textExpression} == ${expected}`;
    if (rule.operator === 'neq') return `${textExpression} != ${expected}`;
    if (rule.operator === 'contains') return `${expected} in ${textExpression}`;
    if (rule.operator === 'not_contains') return `${expected} not in ${textExpression}`;
    if (rule.operator === 'empty') return `not ${textExpression}`;
    if (rule.operator === 'not_empty') return `${textExpression}`;
    if (typeof expectedNumber === 'number' && !Number.isFinite(expectedNumber)) return null;
    if (rule.operator === 'lt') return `${numberExpression} < ${expectedNumber}`;
    if (rule.operator === 'lte') return `${numberExpression} <= ${expectedNumber}`;
    if (rule.operator === 'gt') return `${numberExpression} > ${expectedNumber}`;
    if (rule.operator === 'gte') return `${numberExpression} >= ${expectedNumber}`;
    if (rule.operator === 'between') {
      if (typeof expectedMaxNumber === 'number' && !Number.isFinite(expectedMaxNumber)) return null;
      return `${numberExpression} >= ${expectedNumber} and ${numberExpression} <= ${expectedMaxNumber}`;
    }
    return null;
  });
  const setup = rules.map((rule, index) => {
    const valueSetup = rule.value.includes('{{') || rule.value.includes('{%')
      ? `{% set statusExpected${index} %}${statusRuleValueTemplate(rule.value)}{% endset %}`
      : '';
    const valueMaxSetup = rule.valueMax.includes('{{') || rule.valueMax.includes('{%')
      ? `{% set statusExpectedMax${index} %}${statusRuleValueTemplate(rule.valueMax)}{% endset %}`
      : '';
    return `${valueSetup}${valueMaxSetup}`;
  }).join('');
  const branches = rules
    .map((rule, index) => ({ rule, condition: conditions[index] }))
    .filter(({ condition }) => condition)
    .map(({ rule, condition }, index) => `${index === 0 ? 'if' : 'elif'} ${condition} %}${accentCssColor(rule.accent)}{% `)
    .join('');

  return branches
    ? `${setup}{% ${branches}else %}${accentCssColor(fallbackAccent)}{% endif %}`
    : accentCssColor(fallbackAccent);
}

function templateExpression(value, fallback) {
  const normalized = String(value ?? '').trim();
  return normalized || fallback;
}

function dataPath(value, fallback) {
  return String(value ?? fallback).trim() || fallback;
}

function stripItemPrefix(value) {
  return dataPath(value, '').replace(/^item\./, '');
}

function unwrapOutputExpression(value) {
  const normalized = dataPath(value, '');
  const match = normalized.match(/^\{\{\s*([\s\S]*?)\s*\}\}$/);
  return match ? match[1].trim() : '';
}

function hasTemplateSyntax(value) {
  const normalized = dataPath(value, '');
  return normalized.includes('{{') || normalized.includes('{%');
}

function itemExpression(field, scalarFallback = 'value') {
  const outputExpression = unwrapOutputExpression(field);
  if (outputExpression) return outputExpression;
  const normalized = stripItemPrefix(field);
  if (!normalized) return scalarFallback;
  if (normalized === 'key') return `(key if key is defined else ${scalarFallback})`;
  if (normalized === 'value') return `(_dcRowValue if _dcRowValue is defined else (value if value is defined else item))`;
  return `(item.${normalized} if item is mapping else ${scalarFallback})`;
}

function itemOutput(field, scalarFallback = 'value') {
  if (hasTemplateSyntax(field)) return dataPath(field, '');
  return `{{ ${itemExpression(field, scalarFallback)} }}`;
}

function itemFieldExpression(field, fallbackField) {
  return itemExpression(dataPath(field, fallbackField), 'value');
}

function collectionTemplate(items, arrayBody, objectBody, emptyBody) {
  return `{% if ${items} and ${items}.length %}{% for item in ${items} %}{% set key = loop.index0 %}{% set value = item %}{% set _dcRowValue = item %}${arrayBody}{% endfor %}{% elif ${items} %}{% for key, item in ${items} %}{% set value = item %}{% set _dcRowValue = item %}${objectBody}{% else %}${emptyBody}{% endfor %}{% else %}${emptyBody}{% endif %}`;
}

function booleanOption(value, fallback = 'false') {
  const normalized = String(value ?? fallback).trim().toLowerCase();
  return normalized === 'true' ? 'true' : 'false';
}

function toBoolean(value) {
  return String(value).trim().toLowerCase() === 'true';
}

function integerOption(value, fallback, min = 0) {
  const number = Number.parseInt(value, 10);
  return Number.isFinite(number) && number >= min ? number : fallback;
}

function hashString(value) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = ((hash << 5) - hash) + value.charCodeAt(index);
    hash |= 0;
  }
  return Math.abs(hash).toString(36);
}

const progressCard = {
  key: 'template/progress-card',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  name: 'Progression',
  description: 'Affiche un objectif, sa valeur actuelle et une jauge de progression.',
  preview: '<div class="d-flex justify-content-between"><div><div class="fw-semibold">Objectif mensuel</div><div class="small text-secondary">75 / 100 dossiers</div></div><div class="fw-semibold">75%</div></div><div class="progress mt-3" role="progressbar" aria-valuenow="75" aria-valuemin="0" aria-valuemax="100"><div class="progress-bar bg-info" style="width: 75%"></div></div>',
  fields: [
    { key: 'title', type: 'text', required: true, placeholder: 'Objectif mensuel', help: 'Nom de l’objectif affiché.' },
    { key: 'current', type: 'template', required: true, placeholder: '{{ done }}', help: 'Valeur dynamique actuelle.' },
    { key: 'target', type: 'template', required: true, placeholder: '{{ target }}', help: 'Valeur dynamique cible.' },
    { key: 'percent', type: 'template', required: true, placeholder: '{{ percent }}', help: 'Pourcentage dynamique entre 0 et 100.' },
    { key: 'unit', type: 'text', placeholder: 'dossiers', help: 'Unité optionnelle affichée après la cible.' },
    { key: 'accent', type: 'color', options: accentOptions, help: 'Couleur utilisée si aucune règle de seuil ne correspond.' },
    { key: 'colorRules', type: 'threshold-colors', options: accentOptions, help: 'Règles de couleur évaluées dans l’ordre sur le pourcentage.' }
  ],
  defaults: {
    title: 'Objectif',
    current: '',
    target: '',
    percent: '',
    unit: '',
    accent: 'info',
    colorRules: []
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      current: templateExpression(config.current, ''),
      target: templateExpression(config.target, ''),
      percent: templateExpression(config.percent, ''),
      unit: String(config.unit ?? this.defaults.unit),
      accent: normalizeAccent(config.accent, this.defaults.accent),
      colorRules: normalizeProgressColorRules(config.colorRules)
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.current || !normalized.target) {
      return {
        template: `<div class="d-flex justify-content-between gap-3"><div><div class="fw-semibold">${escapeHtml(normalized.title)}</div><div class="small text-secondary text-secondary small">Non configuré</div></div></div>`,
        javascript: ''
      };
    }
    const unit = normalized.unit ? ` ${escapeHtml(normalized.unit)}` : '';
    const progressColor = compileProgressAccentRules(normalized.colorRules, normalized.accent);
    const progressPercentSetup = normalized.colorRules.length
      ? `{% set progressPercentValue %}${normalized.percent || 0}{% endset %}{% set progressPercent = progressPercentValue | float %}`
      : '';

    return {
      template: `${progressPercentSetup}<div class="d-flex justify-content-between gap-3"><div><div class="fw-semibold">${escapeHtml(normalized.title)}</div><div class="small text-secondary">${normalized.current} / ${normalized.target}${unit}</div></div><div class="fw-semibold">${normalized.percent || 0}%</div></div><div class="progress mt-3" role="progressbar" aria-valuenow="${normalized.percent || 0}" aria-valuemin="0" aria-valuemax="100"><div class="progress-bar" style="width: ${normalized.percent || 0}%; background-color: ${progressColor};"></div></div>`,
      javascript: ''
    };
  }
};

const statusList = {
  key: 'template/status-list',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  name: 'Liste de statuts',
  description: 'Liste des éléments avec libellé, détail et badge dynamique.',
  preview: '<ul class="list-group"><li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">API</div><div class="small text-secondary">Latence 42 ms</div></div><span class="badge text-bg-success align-self-center">OK</span></li><li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">Jobs</div><div class="small text-secondary">2 en attente</div></div><span class="badge text-bg-warning align-self-center">À surveiller</span></li></ul>',
  fields: [
    { key: 'title', type: 'text', required: true, placeholder: 'Services', help: 'Titre de la liste.' },
    { key: 'items', type: 'data-path', required: true, placeholder: 'services', help: 'Nom de la liste dans les données. Exemple : services, alerts, clients.' },
    { key: 'labelField', type: 'text', required: true, placeholder: 'name', help: 'Champ utilisé comme libellé principal pour chaque ligne.' },
    { key: 'detailField', type: 'text', placeholder: 'detail', help: 'Champ optionnel affiché sous le libellé.' },
    { key: 'badgeField', type: 'text', required: true, placeholder: 'status', help: 'Champ affiché dans le badge.' },
    { key: 'accent', type: 'color', options: accentOptions, help: 'Couleur utilisée si aucune règle ne correspond.' },
    { key: 'colorRules', type: 'condition-colors', options: accentOptions, help: 'Règles de couleur évaluées dans l’ordre. Le champ comparé est configurable par règle.' }
  ],
  defaults: {
    title: 'Statuts',
    items: '',
    labelField: '',
    detailField: '',
    badgeField: '',
    accent: 'success',
    colorRules: []
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      items: dataPath(config.items, ''),
      labelField: dataPath(config.labelField, ''),
      detailField: String(config.detailField ?? '').trim(),
      badgeField: dataPath(config.badgeField, ''),
      accent: normalizeAccent(config.accent, this.defaults.accent),
      colorRules: normalizeStatusColorRules(config.colorRules)
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.items) {
      return {
        template: `<div class="text-secondary small">Veuillez configurer la source de données (items)</div>`,
        javascript: ''
      };
    }
    const detail = normalized.detailField
      ? `<div class="small text-secondary">${itemOutput(normalized.detailField, 'key')}</div>`
      : '';
    const statusColor = compileStatusAccentRules(normalized.colorRules, normalized.accent, normalized.badgeField);
    const itemHtml = `<li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">${itemOutput(normalized.labelField, 'key')}</div>${detail}</div><span class="badge align-self-center" style="background-color: ${statusColor}; color: #fff;">${itemOutput(normalized.badgeField, 'value')}</span></li>`;
    const emptyHtml = '<li class="list-group-item text-secondary">Aucune donnée</li>';

    return {
      template: `<ul class="list-group">${collectionTemplate(normalized.items, itemHtml, itemHtml, emptyHtml)}</ul>`,
      javascript: ''
    };
  }
};

const compactTable = {
  key: 'template/compact-table',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  name: 'Tableau DataTables',
  description: 'Génère les colonnes depuis une collection et active recherche, tri, pagination, export et options avancées DataTables.',
  preview: '<table class="table table-sm table-striped align-middle w-100"><thead><tr><th>Client</th><th>Segment</th><th>CA</th></tr></thead><tbody><tr><td>Acme</td><td>Enterprise</td><td>18 k€</td></tr><tr><td>Northwind</td><td>Retail</td><td>12 k€</td></tr></tbody></table>',
  fields: [
    { key: 'title', type: 'text', required: true, placeholder: 'Top clients', help: 'Titre du tableau.' },
    { key: 'items', type: 'data-path', required: true, placeholder: 'clients', help: 'Nom de la liste dans les données. Le tableau crée les colonnes automatiquement avec les clés du premier objet.' },
    { key: 'columns', type: 'columns-manager', help: 'Personnalisez et ordonnez les colonnes du tableau, ou laissez vide pour générer automatiquement toutes les colonnes.' },
    { key: 'pageLength', type: 'number', placeholder: '25', help: 'Nombre de lignes affichées par page.' },
    { key: 'searching', type: 'select', options: booleanOptions, help: 'Active le champ de recherche DataTables.' },
    { key: 'ordering', type: 'select', options: booleanOptions, help: 'Active le tri sur les colonnes.' },
    { key: 'paging', type: 'select', options: booleanOptions, help: 'Active la pagination.' },
    { key: 'info', type: 'select', options: booleanOptions, help: 'Affiche les informations de pagination.' },
    { key: 'responsive', type: 'select', options: booleanOptions, help: 'Adapte les colonnes aux petits écrans.' },
    { key: 'buttons', type: 'select', options: dataTableButtonOptions, help: 'Ajoute les boutons de copie, export, impression et visibilité des colonnes.' },
    { key: 'searchBuilder', type: 'select', options: booleanOptions, help: 'Ajoute le constructeur de recherche avancée.' },
    { key: 'searchPanes', type: 'select', options: booleanOptions, help: 'Ajoute les panneaux de filtrage par colonne.' },
    { key: 'colReorder', type: 'select', options: booleanOptions, help: 'Permet de réordonner les colonnes.' },
    { key: 'fixedHeader', type: 'select', options: booleanOptions, help: 'Garde l’en-tête visible pendant le scroll.' },
    { key: 'stateSave', type: 'select', options: booleanOptions, help: 'Mémorise recherche, tri, page et colonnes dans le navigateur.' },
    { key: 'selectRows', type: 'select', options: booleanOptions, help: 'Permet la sélection de lignes.' },
    { key: 'scrollX', type: 'select', options: booleanOptions, help: 'Active le défilement horizontal.' },
    { key: 'scroller', type: 'select', options: booleanOptions, help: 'Optimise l’affichage des grandes listes.' },
    { key: 'fixedColumnsLeft', type: 'number', placeholder: '0', help: 'Nombre de colonnes figées à gauche.' },
    { key: 'fixedColumnsRight', type: 'number', placeholder: '0', help: 'Nombre de colonnes figées à droite.' },
    { key: 'rowGroupColumn', type: 'number', placeholder: '0', help: 'Index de colonne à regrouper, en partant de 1. 0 désactive le regroupement.' },
    { key: 'orderColumn', type: 'number', placeholder: '1', help: 'Index de colonne utilisée pour le tri initial, en partant de 1.' },
    { key: 'orderDirection', type: 'select', options: orderDirectionOptions, help: 'Sens du tri initial.' }
  ],
  defaults: {
    title: 'Tableau',
    items: '',
    columns: '',
    pageLength: 25,
    searching: 'true',
    ordering: 'true',
    paging: 'true',
    info: 'true',
    responsive: 'true',
    buttons: 'basic',
    searchBuilder: 'false',
    searchPanes: 'false',
    colReorder: 'true',
    fixedHeader: 'true',
    stateSave: 'true',
    selectRows: 'false',
    scrollX: 'false',
    scroller: 'false',
    fixedColumnsLeft: 0,
    fixedColumnsRight: 0,
    rowGroupColumn: 0,
    orderColumn: 1,
    orderDirection: 'asc'
  },
  normalizeConfig(config = {}) {
    const buttonValues = dataTableButtonOptions.map((option) => option.value);
    const orderDirection = orderDirectionOptions.some((option) => option.value === config.orderDirection)
      ? config.orderDirection
      : this.defaults.orderDirection;

    return {
      title: String(config.title ?? this.defaults.title),
      items: dataPath(config.items, ''),
      columns: String(config.columns ?? this.defaults.columns),
      pageLength: integerOption(config.pageLength, this.defaults.pageLength, 1),
      searching: booleanOption(config.searching, this.defaults.searching),
      ordering: booleanOption(config.ordering, this.defaults.ordering),
      paging: booleanOption(config.paging, this.defaults.paging),
      info: booleanOption(config.info, this.defaults.info),
      responsive: booleanOption(config.responsive, this.defaults.responsive),
      buttons: buttonValues.includes(config.buttons) ? config.buttons : this.defaults.buttons,
      searchBuilder: booleanOption(config.searchBuilder, this.defaults.searchBuilder),
      searchPanes: booleanOption(config.searchPanes, this.defaults.searchPanes),
      colReorder: booleanOption(config.colReorder, this.defaults.colReorder),
      fixedHeader: booleanOption(config.fixedHeader, this.defaults.fixedHeader),
      stateSave: booleanOption(config.stateSave, this.defaults.stateSave),
      selectRows: booleanOption(config.selectRows, this.defaults.selectRows),
      scrollX: booleanOption(config.scrollX, this.defaults.scrollX),
      scroller: booleanOption(config.scroller, this.defaults.scroller),
      fixedColumnsLeft: integerOption(config.fixedColumnsLeft, this.defaults.fixedColumnsLeft, 0),
      fixedColumnsRight: integerOption(config.fixedColumnsRight, this.defaults.fixedColumnsRight, 0),
      rowGroupColumn: integerOption(config.rowGroupColumn, this.defaults.rowGroupColumn, 0),
      orderColumn: integerOption(config.orderColumn, this.defaults.orderColumn, 1),
      orderDirection
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.items) {
      return {
        template: `<div class="text-secondary small">Veuillez configurer la source de données (items)</div>`,
        javascript: ''
      };
    }
    const buttonSets = {
      none: [],
      basic: ['copy', 'csv', 'excel', 'print'],
      all: ['copy', 'csv', 'excel', 'pdf', 'print', 'colvis']
    };
    const dom = [
      toBoolean(normalized.searchBuilder) ? 'Q' : '',
      toBoolean(normalized.searchPanes) ? 'P' : '',
      normalized.buttons !== 'none' ? 'B' : '',
      'frtip'
    ].join('');
    const orderColumnIndex = Math.max(normalized.orderColumn - 1, 0);
    const rowGroupColumnIndex = normalized.rowGroupColumn > 0 ? normalized.rowGroupColumn - 1 : null;
    const fixedColumns = normalized.fixedColumnsLeft || normalized.fixedColumnsRight
      ? {
           leftColumns: normalized.fixedColumnsLeft,
           rightColumns: normalized.fixedColumnsRight
        }
      : false;
    const dataTableOptions = {
      dom,
      pageLength: normalized.pageLength,
      lengthMenu: [[10, 25, 50, 100, -1], [10, 25, 50, 100, 'All']],
      searching: toBoolean(normalized.searching),
      ordering: toBoolean(normalized.ordering),
      paging: toBoolean(normalized.paging),
      info: toBoolean(normalized.info),
      responsive: toBoolean(normalized.responsive),
      colReorder: toBoolean(normalized.colReorder),
      fixedHeader: toBoolean(normalized.fixedHeader),
      stateSave: toBoolean(normalized.stateSave),
      select: toBoolean(normalized.selectRows),
      scrollX: toBoolean(normalized.scrollX) || Boolean(fixedColumns),
      scroller: toBoolean(normalized.scroller),
      deferRender: toBoolean(normalized.scroller),
      searchBuilder: toBoolean(normalized.searchBuilder),
      searchPanes: toBoolean(normalized.searchPanes),
      fixedColumns,
      rowGroup: rowGroupColumnIndex === null ? false : { dataSrc: rowGroupColumnIndex },
      order: toBoolean(normalized.ordering) ? [[orderColumnIndex, normalized.orderDirection]] : [],
      buttons: buttonSets[normalized.buttons]
    };
    const tableToken = `dc-dt-${hashString(JSON.stringify(normalized))}`;
    const optionsJson = JSON.stringify(dataTableOptions);

    let templateHtml = '';
    let parsedColumns = [];
    if (normalized.columns) {
      try {
        parsedColumns = JSON.parse(normalized.columns);
      } catch (e) {
        parsedColumns = [];
      }
    }

    if (Array.isArray(parsedColumns) && parsedColumns.length > 0) {
      let theadRows = '';
      let tbodyCols = '';

      parsedColumns.forEach((colObj) => {
        const col = colObj.key;
        const headerText = colObj.label || col;
        theadRows += `<th>${headerText}</th>`;
        const valueExpr = itemExpression(col, "''");

        if (colObj.template) {
          tbodyCols += `<td>{% set value = ${valueExpr} %}${colObj.template}</td>`;
        } else {
          tbodyCols += `<td>{{ ${valueExpr} }}</td>`;
        }
      });

      const rowHtml = `<tr>${tbodyCols}</tr>`;
      templateHtml = `{% if ${normalized.items} %}<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}"><thead><tr>${theadRows}</tr></thead><tbody>${collectionTemplate(normalized.items, rowHtml, rowHtml, `<tr><td colspan="${parsedColumns.length}">Aucune donnée</td></tr>`)}</tbody></table>{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    } else {
      const arrayTable = `<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}"><thead><tr>{% for key, value in ${normalized.items}[0] %}<th>{{ key }}</th>{% endfor %}</tr></thead><tbody>{% for item in ${normalized.items} %}<tr>{% for key, value in item %}<td>{{ value }}</td>{% endfor %}</tr>{% endfor %}</tbody></table>`;
      const objectTable = `<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}">{% for key, item in ${normalized.items} %}{% set value = item %}{% if loop.first %}<thead><tr><th>Clé</th>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<th>{{ fieldKey }}</th>{% endfor %}{% else %}<th>Valeur</th>{% endif %}</tr></thead><tbody>{% endif %}<tr><td>{{ key }}</td>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<td>{{ fieldValue }}</td>{% endfor %}{% else %}<td>{{ value }}</td>{% endif %}</tr>{% else %}<tbody><tr><td>Aucune donnée</td></tr>{% endfor %}</tbody></table>`;
      templateHtml = `{% if ${normalized.items} and ${normalized.items}.length %}${arrayTable}{% elif ${normalized.items} %}${objectTable}{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    }

    return {
      template: templateHtml,
      javascript: `const dataTableOptions = ${optionsJson};
if (dataTableOptions.buttons?.includes('excel') && DataTable.Buttons?.jszip) {
  DataTable.Buttons.jszip(jszip);
}
if (dataTableOptions.buttons?.includes('pdf') && DataTable.Buttons?.pdfMake) {
  pdfmake.vfs = pdfFonts?.pdfMake?.vfs || pdfFonts?.vfs || pdfmake.vfs;
  DataTable.Buttons.pdfMake(pdfmake);
}
document.querySelectorAll('table[data-dc-datatable="${tableToken}"]').forEach((table) => {
  if (jQuery.fn?.dataTable?.isDataTable(table)) return;
  new DataTable(table, dataTableOptions);
});`
    };
  }
};

const timeline = {
  key: 'template/timeline',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  name: 'Timeline',
  description: 'Affiche les événements récents sous forme de fil chronologique.',
  preview: '<div class="border-start ps-3"><div class="mb-3"><div class="small text-secondary">09:30</div><div class="fw-medium">Import terminé</div><div class="small text-secondary">1 248 lignes traitées</div></div><div><div class="small text-secondary">10:15</div><div class="fw-medium">Rapport publié</div><div class="small text-secondary">Finance</div></div></div>',
  fields: [
    { key: 'title', type: 'text', required: true, placeholder: 'Activité', help: 'Titre du fil chronologique.' },
    { key: 'items', type: 'data-path', required: true, placeholder: 'events', help: 'Nom de la liste d’événements dans les données.' },
    { key: 'dateField', type: 'text', required: true, placeholder: 'date', help: 'Champ contenant la date ou l’heure.' },
    { key: 'labelField', type: 'text', required: true, placeholder: 'title', help: 'Champ contenant le titre de l’événement.' },
    { key: 'detailField', type: 'text', placeholder: 'detail', help: 'Champ optionnel affiché sous l’événement.' },
    { key: 'accent', type: 'color', options: accentOptions }
  ],
  defaults: {
    title: 'Activité',
    items: '',
    dateField: '',
    labelField: '',
    detailField: '',
    accent: 'primary'
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      items: dataPath(config.items, ''),
      dateField: dataPath(config.dateField, ''),
      labelField: dataPath(config.labelField, ''),
      detailField: String(config.detailField ?? '').trim(),
      accent: normalizeAccent(config.accent, this.defaults.accent)
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.items) {
      return {
        template: `<div class="text-secondary small">Veuillez configurer la source de données (items)</div>`,
        javascript: ''
      };
    }
    const detail = normalized.detailField
      ? `<div class="small text-secondary">${itemOutput(normalized.detailField, "''")}</div>`
      : '';
    const itemHtml = `<div class="mb-3 position-relative"><div class="small text-secondary">${itemOutput(normalized.dateField, 'key')}</div><div class="fw-medium">${itemOutput(normalized.labelField, 'value')}</div>${detail}</div>`;
    const emptyHtml = '<div class="text-secondary">Aucune donnée</div>';

    return {
      template: `<div class="border-start ps-3" style="border-left-color: ${accentCssColor(normalized.accent)} !important;">${collectionTemplate(normalized.items, itemHtml, itemHtml, emptyHtml)}</div>`,
      javascript: ''
    };
  }
};

export const templateCatalog = [
  progressCard,
  statusList,
  compactTable,
  timeline
];

export function getTemplateDefinition(key, major) {
  return templateCatalog.find((template) => template.key === key && template.major === Number(major)) || null;
}
