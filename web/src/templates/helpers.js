export const accentOptions = [
  { value: 'primary', label: 'Bleu' },
  { value: 'success', label: 'Vert' },
  { value: 'info', label: 'Cyan' },
  { value: 'warning', label: 'Jaune' },
  { value: 'danger', label: 'Rouge' },
  { value: 'secondary', label: 'Gris' }
];

export const booleanOptions = [
  { value: 'true', label: 'Oui' },
  { value: 'false', label: 'Non' }
];

export const dataTableButtonOptions = [
  { value: 'none', label: 'Aucun' },
  { value: 'basic', label: 'Export simple' },
  { value: 'all', label: 'Export complet' }
];

export const orderDirectionOptions = [
  { value: 'asc', label: 'Ascendant' },
  { value: 'desc', label: 'Descendant' }
];

const accentValues = accentOptions.map((option) => option.value);
const htmlColorPattern = /^#[0-9a-f]{6}$/i;
const variableNamePattern = /^[A-Za-z_][A-Za-z0-9_]*$/;
const voidHtmlTags = new Set([
  'area',
  'base',
  'br',
  'col',
  'embed',
  'hr',
  'img',
  'input',
  'link',
  'meta',
  'param',
  'source',
  'track',
  'wbr'
]);

export const templateVariableFields = [
  {
    key: 'globalVariables',
    type: 'template-variables',
    section: 'variables',
    scope: 'global',
    placeholder: '{{ items | length }}',
    help: 'Variables Nunjucks évaluées avant le template.'
  },
  {
    key: 'loopVariables',
    type: 'template-variables',
    section: 'variables',
    scope: 'loop',
    placeholder: '{{ item.name if item is mapping else value }}',
    help: 'Variables Nunjucks évaluées au début de chaque boucle. Utilisation : {{ var.loop.nom }}. Elles peuvent utiliser item, value, key et les variables globales.'
  }
];

export function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

export function normalizeAccent(value, fallback = 'primary') {
  if (value && typeof value === 'object' && value.baseField !== undefined) {
    return value;
  }
  const str = String(value ?? '').trim();
  if (str.includes('{') || str.includes('%')) return str;
  if (htmlColorPattern.test(str)) return str;
  return accentValues.includes(str) ? str : fallback;
}

export function accentCssColor(value) {
  const normalized = String(value ?? '').trim();
  if (normalized.includes('{') || normalized.includes('%')) return normalized;
  if (htmlColorPattern.test(normalized)) return normalized;
  return accentValues.includes(normalized) ? `var(--bs-${normalized})` : `var(--bs-primary)`;
}

export function normalizeProgressColorRules(value) {
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

export function compileProgressAccentRules(rules, fallbackAccent) {
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
  if (normalized.includes('{{') || normalized.includes('{%')) return normalized;
  return escapeHtml(normalized);
}

export function normalizeTemplateVariables(value) {
  if (Array.isArray(value)) {
    return value
      .map((variable) => ({
        name: String(variable?.name ?? '').trim(),
        value: String(variable?.value ?? '').trim()
      }))
      .filter((variable) => variableNamePattern.test(variable.name) && variable.value !== '');
  }

  const normalized = String(value ?? '').trim();
  if (!normalized) return [];

  try {
    return normalizeTemplateVariables(JSON.parse(normalized));
  } catch {
    return [];
  }
}

function compileObjectEntries(variableNames, variablePrefix) {
  return variableNames
    .filter((name, index, names) => names.indexOf(name) === index)
    .map((name) => `"${name}": ${variablePrefix}${name}`)
    .join(', ');
}

function uniqueNames(variableNames) {
  return variableNames.filter((name, index, names) => names.indexOf(name) === index);
}

function compileVariableNamespace(globalNames = [], loopNames = []) {
  const uniqueGlobalNames = uniqueNames(globalNames);
  const uniqueLoopNames = uniqueNames(loopNames);
  const globalEntries = compileObjectEntries(uniqueGlobalNames, '_dcGlobal_');
  const loopEntries = compileObjectEntries(uniqueLoopNames, '_dcLoop_');
  const rootEntries = [
    '"global": _dcGlobalVar',
    '"loop": _dcLoopVar'
  ].join(', ');

  return `{% set _dcGlobalVar = {${globalEntries}} %}{% set _dcLoopVar = {${loopEntries}} %}{% set var = {${rootEntries}} %}`;
}

export function compileTemplateVariables(variables, inheritedVariables = [], scope = 'global') {
  const inheritedNames = normalizeTemplateVariables(inheritedVariables).map((variable) => variable.name);
  const declaredNames = [];
  const setup = scope === 'loop' && inheritedNames.length
    ? compileVariableNamespace(inheritedNames, declaredNames)
    : '';
  return setup + normalizeTemplateVariables(variables)
    .map((variable) => {
      declaredNames.push(variable.name);
      const variableReference = scope === 'loop' ? `_dcLoop_${variable.name}` : `_dcGlobal_${variable.name}`;
      const expression = unwrapOutputExpression(variable.value);
      const assignment = expression
        ? `{% set ${variableReference} = ${expression} %}`
        : `{% set ${variableReference} %}${variable.value}{% endset %}`;
      const namespace = scope === 'loop'
        ? compileVariableNamespace(inheritedNames, declaredNames)
        : compileVariableNamespace(declaredNames, []);
      return `${assignment}${namespace}`;
    })
    .join('');
}

export function normalizeStatusColorRules(value) {
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

export function compileStatusAccentRules(rules, fallbackAccent, fallbackField) {
  if (!rules.length) return fallbackAccent;
  const conditions = rules.map((rule, index) => {
    const expectedVar = `statusExpected${index}`;
    const expectedMaxVar = `statusExpectedMax${index}`;
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

export function templateExpression(value, fallback) {
  const normalized = String(value ?? '').trim();
  return normalized || fallback;
}

export function dataPath(value, fallback) {
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

export function itemExpression(field, scalarFallback = 'value') {
  const outputExpression = unwrapOutputExpression(field);
  if (outputExpression) return outputExpression;
  const normalized = stripItemPrefix(field);
  if (!normalized) return scalarFallback;
  if (normalized === 'key') return `(key if key is defined else ${scalarFallback})`;
  if (normalized === 'value') return `(_dcRowValue if _dcRowValue is defined else (value if value is defined else item))`;
  return `(item.${normalized} if item is mapping else ${scalarFallback})`;
}

export function itemOutput(field, scalarFallback = 'value') {
  if (hasTemplateSyntax(field)) return dataPath(field, '');
  return `{{ ${itemExpression(field, scalarFallback)} }}`;
}

export function itemFieldExpression(field, fallbackField) {
  return itemExpression(dataPath(field, fallbackField), 'value');
}

export function collectionTemplate(items, arrayBody, objectBody, emptyBody, loopVariables = [], globalVariables = []) {
  const loopSetup = compileTemplateVariables(loopVariables, globalVariables, 'loop');
  return `{% if ${items} and ${items}.length %}{% for item in ${items} %}{% set key = loop.index0 %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}${arrayBody}{% endfor %}{% elif ${items} %}{% for key, item in ${items} %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}${objectBody}{% else %}${emptyBody}{% endfor %}{% else %}${emptyBody}{% endif %}`;
}

export function booleanOption(value, fallback = 'false') {
  const normalized = String(value ?? fallback).trim().toLowerCase();
  return normalized === 'true' ? 'true' : 'false';
}

export function toBoolean(value) {
  return String(value).trim().toLowerCase() === 'true';
}

export function integerOption(value, fallback, min = 0) {
  const number = Number.parseInt(value, 10);
  return Number.isFinite(number) && number >= min ? number : fallback;
}

export function hashString(value) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = ((hash << 5) - hash) + value.charCodeAt(index);
    hash |= 0;
  }
  return Math.abs(hash).toString(36);
}

function findHtmlTagEnd(template, start) {
  let quote = '';
  for (let index = start + 1; index < template.length; index += 1) {
    const char = template[index];
    if (quote) {
      if (char === quote) quote = '';
      continue;
    }
    if (char === '"' || char === "'") {
      quote = char;
      continue;
    }
    if (char === '>') return index;
  }
  return template.length - 1;
}

function tokenizeTemplate(template) {
  const tokens = [];
  let index = 0;

  while (index < template.length) {
    if (template[index] === '<') {
      const end = findHtmlTagEnd(template, index);
      tokens.push(template.slice(index, end + 1));
      index = end + 1;
      continue;
    }

    const jinjaStart = template.slice(index, index + 2);
    if (['{%', '{#'].includes(jinjaStart)) {
      const closing = jinjaStart === '{%' ? '%}' : '#}';
      const end = template.indexOf(closing, index + 2);
      if (end !== -1) {
        tokens.push(template.slice(index, end + closing.length));
        index = end + closing.length;
        continue;
      }
    }

    const nextHtml = template.indexOf('<', index);
    const nextJinjaStatement = template.indexOf('{%', index);
    const nextJinjaComment = template.indexOf('{#', index);
    const nextIndexes = [nextHtml, nextJinjaStatement, nextJinjaComment].filter((value) => value !== -1);
    const next = nextIndexes.length ? Math.min(...nextIndexes) : template.length;
    tokens.push(template.slice(index, next));
    index = next;
  }

  return tokens.filter((token) => token.trim() !== '');
}

function htmlTagName(token) {
  const match = token.match(/^<\/?\s*([A-Za-z][A-Za-z0-9:-]*)/);
  return match ? match[1].toLowerCase() : '';
}

function isOpeningHtmlTag(token) {
  if (!token.startsWith('<') || token.startsWith('</') || token.startsWith('<!--')) return false;
  const tagName = htmlTagName(token);
  return tagName && !voidHtmlTags.has(tagName) && !/\/\s*>$/.test(token);
}

function isClosingHtmlTag(token) {
  return /^<\/\s*[A-Za-z]/.test(token);
}

function jinjaBlockName(token) {
  const match = token.match(/^\{%-?\s*([A-Za-z_][A-Za-z0-9_]*)/);
  return match ? match[1] : '';
}

function isOpeningJinjaBlock(token) {
  const block = jinjaBlockName(token);
  return ['if', 'for', 'block', 'macro', 'filter', 'call', 'raw', 'with'].includes(block);
}

function isMiddleJinjaBlock(token) {
  return ['elif', 'else'].includes(jinjaBlockName(token));
}

function isClosingJinjaBlock(token) {
  return jinjaBlockName(token).startsWith('end');
}

export function formatTemplateForHumans(template) {
  const tokens = tokenizeTemplate(String(template ?? ''));
  const lines = [];
  let depth = 0;

  tokens.forEach((token) => {
    const normalized = token.trim().replace(/\s+/g, ' ');
    if (!normalized) return;

    if (isClosingHtmlTag(normalized) || isClosingJinjaBlock(normalized) || isMiddleJinjaBlock(normalized)) {
      depth = Math.max(depth - 1, 0);
    }

    lines.push(`${'  '.repeat(depth)}${normalized}`);

    if (isOpeningHtmlTag(normalized) || isOpeningJinjaBlock(normalized) || isMiddleJinjaBlock(normalized)) {
      depth += 1;
    }
  });

  return lines.join('\n');
}

export function validateRequiredFields(template, config = {}) {
  return (template.fields || [])
    .filter((field) => field.required)
    .filter((field) => {
      const value = config[field.key] ?? template.defaults?.[field.key];
      return value === undefined || value === null || String(value).trim() === '';
    })
    .map((field) => ({
      field: field.key,
      message: `Le champ "${field.key}" est obligatoire.`
    }));
}

export function createTemplateDefinition(definition) {
  return {
    migrateConfig(config = {}) {
      return { ...config };
    },
    validateConfig(config = {}) {
      return validateRequiredFields(this, this.normalizeConfig(config));
    },
    ...definition
  };
}
