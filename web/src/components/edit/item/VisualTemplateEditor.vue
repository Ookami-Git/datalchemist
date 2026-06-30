<script setup>
import { computed, inject, ref, watch } from 'vue';
import axios from 'axios';
import {
  getTemplateDefinition,
  validateTemplateConfig
} from '@/templates/catalog.js';
import DataPathPicker from './DataPathPicker.vue';
import DataFieldPicker from './DataFieldPicker.vue';
import DataValuePicker from './DataValuePicker.vue';
import DataTableColumnsManager from './DataTableColumnsManager.vue';
import NunjucksTemplateEditor from './NunjucksTemplateEditor.vue';

const props = defineProps({
  templateMeta: { type: Object, required: true },
  itemId: { type: [String, Number], required: true },
  sourceListVersion: { type: Number, default: 0 }
});

const emit = defineEmits(['update:templateMeta']);

const apiUrl = inject('apiUrl');

const conditionKeyPickerOpen = ref('');
const conditionKeys = ref([]);
const conditionKeysLoading = ref(false);
const conditionKeysError = ref('');
const draggingRuleKey = ref('');
const draggingRuleIndex = ref(null);
const dragOverRuleKey = ref('');
const dragOverRuleIndex = ref(null);
const copiedTemplateVariableKey = ref('');
const detectedDataTableColumnOptions = ref([]);

function isObjectKeyField(field) {
  return field.key.endsWith('Field');
}

function isAbsoluteValueField(field) {
  return field.key === 'value' || field.key === 'prevValue';
}

function hasCollectionField(template) {
  return Boolean(template?.fields?.some((field) => field.key === 'items' && field.type === 'data-path'));
}

const selectedTemplate = computed(() => getTemplateDefinition(
  props.templateMeta.templateKey,
  props.templateMeta.templateMajor
) || null);

const openedSections = ref({
  data: true,
  display: true,
  colors: true,
  variables: false,
  options: false,
  advanced: false
});

const sectionDefinitions = [
  {
    key: 'data',
    icon: 'bi-database',
    title: 'Données',
    description: 'Source, collection et champs à afficher.'
  },
  {
    key: 'display',
    icon: 'bi-layout-text-window',
    title: 'Affichage',
    description: 'Libellés, formats et options visibles.'
  },
  {
    key: 'colors',
    icon: 'bi-palette',
    title: 'Couleurs',
    description: 'Accent et règles conditionnelles.'
  },
  {
    key: 'options',
    icon: 'bi-toggle-on',
    title: 'Options & Fonctionnalités',
    description: 'Réglages courants activables rapidement.'
  },
  {
    key: 'advanced',
    icon: 'bi-sliders2',
    title: 'Avancé',
    description: 'Exports, persistance, colonnes figées et fonctions DataTables.'
  },
  {
    key: 'variables',
    icon: 'bi-braces',
    title: 'Variables',
    description: 'Calculs réutilisables dans les expressions du template.'
  }
];

function updateField(key, value) {
  const nextConfig = { ...props.templateMeta.config, [key]: value };
  const field = fields.value.find((candidate) => candidate.key === key);
  if (String(value) === 'true' && field?.exclusiveWith) {
    const exclusiveKeys = Array.isArray(field.exclusiveWith) ? field.exclusiveWith : [field.exclusiveWith];
    exclusiveKeys.forEach((exclusiveKey) => {
      nextConfig[exclusiveKey] = 'false';
    });
  }

  emit('update:templateMeta', {
    ...props.templateMeta,
    config: nextConfig
  });
}

function optionValue(option) {
  return typeof option === 'object' ? option.value : option;
}

function optionLabel(option) {
  return typeof option === 'object' ? option.label : option;
}

function fieldPlaceholder(field) {
  if (field.placeholder) return field.placeholder;
  if (field.type === 'template') return props.templateMeta.config?.[field.key] || '{{ value }}';
  if (field.type === 'data-path') return 'items';
  return '';
}

function isBooleanField(field) {
  return field.type === 'select' && 
         Array.isArray(field.options) && 
         field.options.length === 2 && 
         field.options.some(o => optionValue(o) === 'true') && 
         field.options.some(o => optionValue(o) === 'false');
}

function getConfiguredFieldValue(key) {
  return props.templateMeta.config?.[key] ?? selectedTemplate.value?.defaults?.[key];
}

function isFieldEnabled(key) {
  return String(getConfiguredFieldValue(key)) === 'true';
}

function fieldDependencyMet(field, visited = new Set()) {
  if (!field.dependsOn) return true;
  if (Array.isArray(field.dependsOn)) {
    return field.dependsOn.every((dependency) => dependencyMet(dependency, visited));
  }
  return dependencyMet(field.dependsOn, visited);
}

function dependencyMet(dependency, visited = new Set()) {
  const dependencyKey = typeof dependency === 'string' ? dependency : dependency?.field;
  if (!dependencyKey) return true;

  const dependencyField = fields.value.find((field) => field.key === dependencyKey);
  if (dependencyField?.dependsOn && !visited.has(dependencyKey)) {
    const nextVisited = new Set(visited);
    nextVisited.add(dependencyKey);
    if (!fieldDependencyMet(dependencyField, nextVisited)) return false;
  }

  if (typeof dependency === 'string') return isFieldEnabled(dependency);

  const value = getConfiguredFieldValue(dependency.field);
  return String(value) === String(dependency.value ?? 'true');
}

function groupTitle(groupKey) {
  return selectedTemplate.value?.fieldGroups?.[groupKey] || groupKey;
}

function groupFields(fieldsToGroup) {
  const groups = [];

  fieldsToGroup.forEach((field) => {
    const key = field.group || '_default';
    let group = groups.find((entry) => entry.key === key);
    if (!group) {
      group = {
        key,
        title: key === '_default' ? '' : groupTitle(key),
        fields: []
      };
      groups.push(group);
    }
    group.fields.push(field);
  });

  return groups;
}

const fields = computed(() => selectedTemplate.value?.fields || []);
const selectedTemplateHasCollection = computed(() => hasCollectionField(selectedTemplate.value));
const visibleFields = computed(() => fields.value.filter((field) => fieldDependencyMet(field)));
const templateItemsPath = computed(() => props.templateMeta.config?.items ?? '');
const validationIssues = computed(() => (
  selectedTemplate.value
    ? validateTemplateConfig(selectedTemplate.value, props.templateMeta.config || {})
    : []
));
const sections = computed(() => sectionDefinitions
  .map((section) => {
    const sectionFields = visibleFields.value.filter((field) => (field.section || 'display') === section.key);
    return {
      ...section,
      fields: sectionFields,
      fieldGroups: groupFields(sectionFields)
    };
  })
  .filter((section) => section.fieldGroups.length));

function toggleSection(key) {
  openedSections.value = {
    ...openedSections.value,
    [key]: !openedSections.value[key]
  };
}

function fieldIssue(field) {
  return validationIssues.value.find((issue) => issue.field === field.key) || null;
}

function getFieldClass(field) {
  if (field.type === 'columns-manager') return 'col-12';
  if (field.type === 'condition-colors') return 'col-12';
  if (field.type === 'threshold-colors') return 'col-12';
  if (field.type === 'template-variables') return 'col-12';
  if (field.type === 'template') return 'col-12';
  if (field.key === 'title' || field.key === 'items') return 'col-12 col-md-6';
  if (field.type === 'range') return 'col-12 col-md-6';
  if (field.type === 'number') return 'col-6 col-md-4';
  if (field.type === 'datatable-columns-select') return 'col-12';
  if (field.type === 'datatable-column-select') return 'col-12 col-md-6';
  if (field.type === 'select') return 'col-12 col-md-6';
  if (field.type === 'color') return 'col-12';
  return 'col-12 col-md-6';
}

function isFieldValueTrue(field) {
  const val = getConfiguredFieldValue(field.key);
  if (val === undefined || val === null) {
    return selectedTemplate.value?.defaults?.[field.key] === 'true';
  }
  return String(val) === 'true';
}

function toggleBooleanField(field) {
  const currentValue = isFieldValueTrue(field);
  updateField(field.key, currentValue ? 'false' : 'true');
}

function getFieldDisplayValue(field) {
  const val = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? '';
  return typeof val === 'object' ? '' : val;
}

function configuredDataTableColumnOptions() {
  const rawColumns = props.templateMeta.config?.columns ?? selectedTemplate.value?.defaults?.columns ?? '';
  if (!rawColumns) return [];

  try {
    const parsed = JSON.parse(rawColumns);
    return Array.isArray(parsed)
      ? parsed
          .map((column) => ({
            value: String(column?.key ?? '').trim(),
            label: String(column?.label || column?.key || '').trim()
          }))
          .filter((column) => column.value)
      : [];
  } catch {
    return [];
  }
}

function dataTableColumnOptionsFromData(data, subPath = '') {
  let target = subPath ? getValueByPath(data, subPath) : data;
  const isArrayCollection = Array.isArray(target);

  if (Array.isArray(target)) {
    target = target.length > 0 ? target[0] : null;
  } else if (target && typeof target === 'object' && !isArrayCollection) {
    const values = Object.values(target);
    const hasOnlyScalarValues = values.length > 0 && values.every(value => (
      value === null || typeof value !== 'object'
    ));
    const firstObjectValue = values.find(value => value && typeof value === 'object' && !Array.isArray(value));

    if (hasOnlyScalarValues) {
      return [
        { value: 'key', label: 'Clé' },
        { value: 'value', label: 'Valeur' }
      ];
    }

    target = firstObjectValue || target;
  }

  return target && typeof target === 'object'
    ? Object.keys(target).map((key) => ({ value: key, label: key }))
    : [];
}

async function loadDataTableColumnOptions(pathVal) {
  detectedDataTableColumnOptions.value = [];
  if (!pathVal || configuredDataTableColumnOptions().length > 0) return;

  try {
    const sourcesRes = await axios.get(`${apiUrl}/item/sources/${props.itemId}`);
    const activeSources = sourcesRes.data || [];
    let sourceName = '';
    let sourceId = null;
    let subPath = '';

    if (pathVal.startsWith('sn.')) {
      const parts = pathVal.slice(3).split('.');
      sourceName = parts[0];
      if (sourceName.startsWith("['") || sourceName.startsWith("[\"")) {
        sourceName = sourceName.slice(2, -2);
      }
      subPath = parts.slice(1).join('.');
    } else if (pathVal.startsWith('sid.s')) {
      const parts = pathVal.slice(5).split('.');
      sourceId = parseInt(parts[0], 10);
      subPath = parts.slice(1).join('.');
    } else {
      return;
    }

    const source = sourceId
      ? activeSources.find((candidate) => candidate.id === sourceId)
      : activeSources.find((candidate) => candidate.name === sourceName);

    if (!source) return;

    let sourceDefaults = {};
    try {
      const sourceRes = await axios.get(`${apiUrl}/source/${source.id}`);
      sourceDefaults = sourceRes.data?.json ? (JSON.parse(sourceRes.data.json)?.getDefaults || {}) : {};
    } catch {
      sourceDefaults = {};
    }

    const res = await axios.get(`${apiUrl}/data/source/${source.id}`, {
      params: sourceDefaults
    });
    detectedDataTableColumnOptions.value = dataTableColumnOptionsFromData(res.data, subPath);
  } catch (error) {
    detectedDataTableColumnOptions.value = [];
    console.error('Unable to detect DataTables columns', error);
  }
}

function dataTableColumnOptions(field) {
  const configuredOptions = configuredDataTableColumnOptions();
  const options = configuredOptions.length > 0 ? configuredOptions : detectedDataTableColumnOptions.value;
  if (field.type === 'datatable-columns-select') {
    return options;
  }

  const currentValue = getFieldDisplayValue(field);

  if (options.some((option) => option.value === currentValue)) {
    return options;
  }

  if (/^[1-9]\d*$/.test(String(currentValue))) {
    const legacyIndex = Number.parseInt(currentValue, 10) - 1;
    if (options[legacyIndex]) {
      return options;
    }
    return [
      ...options,
      { value: currentValue, label: `Colonne ${currentValue}` }
    ];
  }

  return currentValue
    ? [...options, { value: currentValue, label: currentValue }]
    : options;
}

function getDataTableColumnSelectValue(field) {
  const currentValue = getFieldDisplayValue(field);
  if (!currentValue || currentValue === '0' || currentValue === 'null' || currentValue === 'undefined') {
    return '';
  }
  const configuredOptions = configuredDataTableColumnOptions();
  const options = configuredOptions.length > 0 ? configuredOptions : detectedDataTableColumnOptions.value;

  if (options.some((option) => option.value === currentValue)) {
    return currentValue;
  }

  if (/^[1-9]\d*$/.test(String(currentValue))) {
    const legacyOption = options[Number.parseInt(currentValue, 10) - 1];
    return legacyOption?.value ?? currentValue;
  }

  return currentValue;
}

function getDataTableColumnsSelectValue(field) {
  const rawValue = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? [];
  if (Array.isArray(rawValue)) {
    return rawValue.map((value) => String(value ?? '').trim()).filter(Boolean);
  }

  const normalized = String(rawValue ?? '').trim();
  if (!normalized) return [];

  try {
    const parsed = JSON.parse(normalized);
    return Array.isArray(parsed)
      ? parsed.map((value) => String(value ?? '').trim()).filter(Boolean)
      : [];
  } catch {
    return normalized.split(',').map((value) => value.trim()).filter(Boolean);
  }
}

function isDataTableColumnSelected(field, value) {
  return getDataTableColumnsSelectValue(field).includes(value);
}

function toggleDataTableColumnSelection(field, value) {
  const currentValues = getDataTableColumnsSelectValue(field);
  const nextValues = currentValues.includes(value)
    ? currentValues.filter((currentValue) => currentValue !== value)
    : [...currentValues, value];
  updateField(field.key, JSON.stringify(nextValues));
}

function getRangeFieldValue(field) {
  const value = Number.parseFloat(getFieldDisplayValue(field));
  if (Number.isFinite(value)) return value;
  return Number.parseFloat(selectedTemplate.value?.defaults?.[field.key] ?? field.min ?? 0) || 0;
}

watch([templateItemsPath, () => props.templateMeta.config?.columns], ([pathVal]) => {
  loadDataTableColumnOptions(pathVal);
}, { immediate: true });

function templateVariablesValue(field) {
  const value = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? [];
  if (Array.isArray(value)) {
    return value.map((variable) => ({
      name: String(variable?.name ?? ''),
      value: String(variable?.value ?? '')
    }));
  }

  const normalized = String(value ?? '').trim();
  if (!normalized) return [];

  try {
    const parsed = JSON.parse(normalized);
    return Array.isArray(parsed)
      ? parsed.map((variable) => ({
          name: String(variable?.name ?? ''),
          value: String(variable?.value ?? '')
        }))
      : [];
  } catch {
    return [];
  }
}

function addTemplateVariable(field) {
  updateField(field.key, [
    ...templateVariablesValue(field),
    { name: '', value: '' }
  ]);
}

function updateTemplateVariable(field, variableIndex, key, value) {
  const variables = templateVariablesValue(field);
  variables[variableIndex] = {
    ...variables[variableIndex],
    [key]: value
  };
  updateField(field.key, variables);
}

function removeTemplateVariable(field, variableIndex) {
  updateField(field.key, templateVariablesValue(field).filter((_, index) => index !== variableIndex));
}

function validTemplateVariableName(name) {
  return /^[A-Za-z_][A-Za-z0-9_]*$/.test(String(name ?? '').trim());
}

function templateVariableToken(field, variable) {
  const name = String(variable?.name ?? '').trim();
  if (!validTemplateVariableName(name)) return '';
  return field.scope === 'loop' ? `{{ var.loop.${name} }}` : `{{ var.global.${name} }}`;
}

async function copyTemplateVariableToken(tokenKey, token) {
  if (!token) return;

  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(token);
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = token;
      textArea.setAttribute('readonly', '');
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }

    copiedTemplateVariableKey.value = tokenKey;
    window.setTimeout(() => {
      if (copiedTemplateVariableKey.value === tokenKey) {
        copiedTemplateVariableKey.value = '';
      }
    }, 1500);
  } catch (error) {
    console.error('Unable to copy template variable token', error);
  }
}

function appendTemplateToken(field, token) {
  const current = getFieldDisplayValue(field);
  updateField(field.key, current ? `${current} ${token}` : token);
}

function isHtmlColor(value) {
  return /^#[0-9a-f]{6}$/i.test(String(value ?? '').trim());
}

function getColorValue(field, fallback = '#0d6efd') {
  const value = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? '';
  return isHtmlColor(value) ? value : fallback;
}

function getRuleColorValue(rule, fallback = '#0d6efd') {
  return isHtmlColor(rule?.accent) ? rule.accent : fallback;
}

function getRuleAccentSelectValue(rule) {
  return isHtmlColor(rule?.accent) ? 'custom' : (rule?.accent || '');
}

function updateRuleAccent(field, ruleIndex, value, fallback) {
  const accent = value === 'custom' ? getRuleColorValue(colorRulesValue(field)[ruleIndex], fallback) : value;
  if (field.type === 'threshold-colors') {
    updateThresholdRule(field, ruleIndex, 'accent', accent);
  } else if (field.type === 'condition-colors') {
    updateConditionRule(field, ruleIndex, 'accent', accent);
  }
}

function thresholdRulesValue(field) {
  const value = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? [];
  return Array.isArray(value) ? value : [];
}

function updateThresholdRule(field, index, key, value) {
  const rules = thresholdRulesValue(field).map((rule) => ({ ...rule }));
  rules[index] = { ...rules[index], [key]: value };
  updateField(field.key, rules);
}

function addThresholdRule(field) {
  updateField(field.key, [
    ...thresholdRulesValue(field),
    { operator: 'lt', value: '', valueMax: '', accent: 'warning' }
  ]);
}

function removeThresholdRule(field, index) {
  updateField(field.key, thresholdRulesValue(field).filter((_, ruleIndex) => ruleIndex !== index));
}

function conditionRulesValue(field) {
  const value = props.templateMeta.config?.[field.key] ?? selectedTemplate.value?.defaults?.[field.key] ?? [];
  return Array.isArray(value) ? value : [];
}

function updateConditionRule(field, index, key, value) {
  const rules = conditionRulesValue(field).map((rule) => ({ ...rule }));
  rules[index] = { ...rules[index], [key]: value };
  updateField(field.key, rules);
}

function addConditionRule(field) {
  updateField(field.key, [
    ...conditionRulesValue(field),
    { field: '', operator: 'contains', value: '', valueMax: '', accent: 'warning' }
  ]);
}

function removeConditionRule(field, index) {
  updateField(field.key, conditionRulesValue(field).filter((_, ruleIndex) => ruleIndex !== index));
}

function colorRulesValue(field) {
  if (field.type === 'threshold-colors') return thresholdRulesValue(field);
  if (field.type === 'condition-colors') return conditionRulesValue(field);
  return [];
}

function clearRuleDragState() {
  draggingRuleKey.value = '';
  draggingRuleIndex.value = null;
  dragOverRuleKey.value = '';
  dragOverRuleIndex.value = null;
}

function onRuleDragStart(field, ruleIndex, event) {
  clearRuleDragState();
  draggingRuleKey.value = field.key;
  draggingRuleIndex.value = ruleIndex;
  event.dataTransfer.effectAllowed = 'move';
  event.dataTransfer.setData('text/plain', `${field.key}:${ruleIndex}`);
}

function onRuleDragOver(field, ruleIndex, event) {
  if (draggingRuleKey.value !== field.key) return;
  if (draggingRuleIndex.value === ruleIndex) return;
  event.preventDefault();
  dragOverRuleKey.value = field.key;
  dragOverRuleIndex.value = ruleIndex;
}

function onRuleDragLeave(field, ruleIndex) {
  if (dragOverRuleKey.value === field.key && dragOverRuleIndex.value === ruleIndex) {
    dragOverRuleKey.value = '';
    dragOverRuleIndex.value = null;
  }
}

function onRuleDrop(field, ruleIndex) {
  if (draggingRuleKey.value !== field.key || draggingRuleIndex.value === null) return;
  if (draggingRuleIndex.value === ruleIndex) {
    clearRuleDragState();
    return;
  }

  const rules = colorRulesValue(field).map((rule) => ({ ...rule }));
  const [movedRule] = rules.splice(draggingRuleIndex.value, 1);
  rules.splice(ruleIndex, 0, movedRule);
  updateField(field.key, rules);
  clearRuleDragState();
}

function ruleDragClass(field, ruleIndex) {
  return {
    dragging: draggingRuleKey.value === field.key && draggingRuleIndex.value === ruleIndex,
    'drag-over-before': draggingRuleKey.value === field.key && dragOverRuleKey.value === field.key && dragOverRuleIndex.value === ruleIndex && draggingRuleIndex.value > ruleIndex,
    'drag-over-after': draggingRuleKey.value === field.key && dragOverRuleKey.value === field.key && dragOverRuleIndex.value === ruleIndex && draggingRuleIndex.value < ruleIndex
  };
}

function getValueByPath(obj, path) {
  if (!path || path === 'racine') return obj;
  const parts = path.split('.');
  let current = obj;
  for (const part of parts) {
    if (current === null || current === undefined) return undefined;

    if (part.includes('[') && part.includes(']')) {
      const cleanPart = part.substring(0, part.indexOf('['));
      const index = parseInt(part.substring(part.indexOf('[') + 1, part.indexOf(']')), 10);
      current = current[cleanPart];
      if (Array.isArray(current)) {
        current = current[index];
      } else {
        return undefined;
      }
    } else {
      current = current[part];
    }
  }
  return current;
}

function extractKeysFromData(data, subPath = '') {
  let target = subPath ? getValueByPath(data, subPath) : data;
  if (Array.isArray(target)) {
    target = target.length > 0 ? target[0] : null;
  }
  conditionKeys.value = typeof target === 'object' && target !== null ? Object.keys(target) : [];
}

async function loadConditionKeys() {
  conditionKeys.value = [];
  conditionKeysError.value = '';
  if (!templateItemsPath.value) {
    conditionKeysError.value = 'Configurez d’abord la collection.';
    return;
  }

  conditionKeysLoading.value = true;
  try {
    const sourcesRes = await axios.get(`${apiUrl}/item/sources/${props.itemId}`);
    const activeSources = sourcesRes.data || [];
    let sourceName = '';
    let sourceId = null;
    let subPath = '';

    if (templateItemsPath.value.startsWith('sn.')) {
      const parts = templateItemsPath.value.slice(3).split('.');
      sourceName = parts[0];
      if (sourceName.startsWith("['") || sourceName.startsWith("[\"")) {
        sourceName = sourceName.slice(2, -2);
      }
      subPath = parts.slice(1).join('.');
    } else if (templateItemsPath.value.startsWith('sid.s')) {
      const parts = templateItemsPath.value.slice(5).split('.');
      sourceId = parseInt(parts[0], 10);
      subPath = parts.slice(1).join('.');
    } else {
      conditionKeysError.value = 'Collection non liée à une source détectable.';
      return;
    }

    const source = sourceId
      ? activeSources.find((candidate) => candidate.id === sourceId)
      : activeSources.find((candidate) => candidate.name === sourceName);

    if (!source) {
      conditionKeysError.value = 'Source introuvable ou non associée à l’item.';
      return;
    }

    let sourceDefaults = {};
    try {
      const sourceRes = await axios.get(`${apiUrl}/source/${source.id}`);
      sourceDefaults = sourceRes.data?.json ? (JSON.parse(sourceRes.data.json)?.getDefaults || {}) : {};
    } catch {
      sourceDefaults = {};
    }

    const res = await axios.get(`${apiUrl}/data/source/${source.id}`, {
      params: sourceDefaults
    });
    extractKeysFromData(res.data, subPath);
  } catch (err) {
    conditionKeysError.value = 'Erreur de chargement des clés.';
    console.error(err);
  } finally {
    conditionKeysLoading.value = false;
  }
}

async function toggleConditionKeyPicker(field, ruleIndex) {
  const pickerKey = `${field.key}:${ruleIndex}`;
  if (conditionKeyPickerOpen.value === pickerKey) {
    conditionKeyPickerOpen.value = '';
    return;
  }
  conditionKeyPickerOpen.value = pickerKey;
  await loadConditionKeys();
}

function itemFieldTemplate(keyName) {
  if (/^[A-Za-z_$][A-Za-z0-9_$]*$/.test(keyName)) {
    return `{{ item.${keyName} }}`;
  }
  return `{{ item[${JSON.stringify(keyName)}] }}`;
}

function selectConditionKey(field, ruleIndex, keyName) {
  updateConditionRule(field, ruleIndex, 'field', itemFieldTemplate(keyName));
  conditionKeyPickerOpen.value = '';
}
</script>

<template>
  <article class="card admin-edit-item-panel admin-edit-item-editor-panel shadow-sm">
    <div class="card-header d-flex align-items-center gap-2">
      <i class="bi bi-sliders"></i>
      <span class="fw-semibold">{{ $t('edititem.visual.configuration') }}</span>
    </div>
    <div class="card-body p-3 p-lg-4">
      <div v-if="selectedTemplate" class="visual-template-workspace">
        <div class="visual-template-builder d-flex flex-column gap-3">
          <div v-if="validationIssues.length" class="alert alert-warning py-2 px-3 mb-0 small">
            <div class="fw-semibold mb-1">Configuration à compléter</div>
            <div v-for="issue in validationIssues" :key="issue.field">{{ issue.message }}</div>
          </div>

          <section
            v-for="section in sections"
            :key="section.key"
            :class="['section-' + section.key, { 'is-open': openedSections[section.key] }]"
            class="visual-template-section border rounded-3 bg-body"
          >
            <button
              type="button"
              class="visual-template-section-toggle w-100 border-0 bg-transparent p-3 d-flex align-items-center gap-3 text-start"
              @click="toggleSection(section.key)"
            >
              <span class="visual-template-section-icon rounded-2 d-inline-flex align-items-center justify-content-center">
                <i class="bi" :class="section.icon"></i>
              </span>
              <span class="flex-grow-1 min-w-0">
                <span class="d-block fw-semibold">{{ section.title }}</span>
                <span class="d-block small text-secondary">{{ section.description }}</span>
              </span>
              <span v-if="section.fields.some(fieldIssue)" class="badge text-bg-warning">À compléter</span>
              <i class="bi" :class="openedSections[section.key] ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
            </button>

            <div v-if="openedSections[section.key]" class="px-3 pb-3">
              <div
                v-for="group in section.fieldGroups"
                :key="group.key"
                class="visual-template-field-group"
              >
                <div v-if="group.title" class="visual-template-field-group-title">{{ group.title }}</div>
                <div class="row g-3">
          <div 
            v-for="field in group.fields" 
            :key="field.key" 
            :class="isBooleanField(field) ? 'col-12 col-md-6 col-lg-4' : getFieldClass(field)"
          >
            <div
              v-if="isBooleanField(field)"
              class="visual-template-switch-card p-3 rounded-3 border h-100 d-flex align-items-start gap-3 transition-all cursor-pointer"
              :class="{ active: isFieldValueTrue(field) }"
              @click="toggleBooleanField(field)"
            >
              <div class="form-check form-switch m-0 pt-0.5">
                <input
                  class="form-check-input pe-none"
                  type="checkbox"
                  role="switch"
                  :checked="isFieldValueTrue(field)"
                  readonly
                />
              </div>
              <div class="d-flex flex-column min-w-0 lh-sm">
                <span class="fw-semibold text-emphasis-dark mb-1" style="font-size: 0.88rem;">
                  {{ $t(`edititem.visual.fields.${field.key}`) }}
                </span>
                <span v-if="field.help" class="text-secondary" style="font-size: 0.75rem; line-height: 1.3;">
                  {{ field.help }}
                </span>
              </div>
            </div>

            <div v-else class="visual-template-field h-100 d-flex flex-column justify-content-between">
              <div>
                <label v-if="field.type !== 'columns-manager'" :for="`visual-${field.key}`" class="form-label visual-template-label mb-2 d-flex justify-content-between align-items-center w-100">
                  <span>
                    <span>{{ $t(`edititem.visual.fields.${field.key}`) }}</span>
                    <span v-if="field.required" class="text-danger">*</span>
                  </span>
                </label>
                
                <select v-if="field.type === 'select'" :id="`visual-${field.key}`" class="form-select shadow-xs"
                  :value="templateMeta.config?.[field.key] ?? selectedTemplate.defaults?.[field.key]" 
                  @change="updateField(field.key, $event.target.value)">
                  <option v-for="option in field.options" :key="optionValue(option)" :value="optionValue(option)">
                    {{ optionLabel(option) }}
                  </option>
                </select>

                <select
                  v-else-if="field.type === 'datatable-column-select'"
                  :id="`visual-${field.key}`"
                  class="form-select shadow-xs"
                  :value="getDataTableColumnSelectValue(field)"
                  :disabled="dataTableColumnOptions(field).length === 0"
                  @change="updateField(field.key, $event.target.value)"
                >
                  <option v-if="dataTableColumnOptions(field).length === 0" value="">
                    Configurez les colonnes du tableau
                  </option>
                  <template v-else>
                    <option v-if="field.allowEmpty || field.placeholder" value="">
                      {{ field.placeholder || 'Aucun' }}
                    </option>
                    <option
                      v-for="option in dataTableColumnOptions(field)"
                      :key="option.value"
                      :value="option.value"
                    >
                      {{ option.label }}
                    </option>
                  </template>
                </select>

                <div
                  v-else-if="field.type === 'datatable-columns-select'"
                  :id="`visual-${field.key}`"
                  class="d-flex flex-column gap-2"
                >
                  <div v-if="dataTableColumnOptions(field).length === 0" class="text-secondary small">
                    Configurez les colonnes du tableau
                  </div>
                  <label
                    v-for="option in dataTableColumnOptions(field)"
                    :key="option.value"
                    class="d-flex align-items-center gap-2 border rounded px-3 py-2 bg-body-tertiary"
                  >
                    <input
                      class="form-check-input m-0"
                      type="checkbox"
                      :checked="isDataTableColumnSelected(field, option.value)"
                      @change="toggleDataTableColumnSelection(field, option.value)"
                    >
                    <span class="small fw-medium">{{ option.label }}</span>
                  </label>
                </div>
                
                <div v-else-if="field.type === 'color'" class="visual-template-color-grid" role="radiogroup"
                  :aria-label="$t(`edititem.visual.fields.${field.key}`)">
                  <button v-for="option in field.options" :key="optionValue(option)" type="button"
                    class="visual-template-color-choice"
                    :class="[`visual-template-color-${optionValue(option)}`, { active: (templateMeta.config?.[field.key] ?? selectedTemplate.defaults?.[field.key]) === optionValue(option) }]"
                    :aria-pressed="(templateMeta.config?.[field.key] ?? selectedTemplate.defaults?.[field.key]) === optionValue(option)"
                    @click="updateField(field.key, optionValue(option))">
                    <span class="visual-template-color-swatch" aria-hidden="true"></span>
                    <span>{{ optionLabel(option) }}</span>
                  </button>
                  <label
                    class="visual-template-color-choice visual-template-color-custom"
                    :class="{ active: isHtmlColor(templateMeta.config?.[field.key] ?? selectedTemplate.defaults?.[field.key]) }"
                  >
                    <input
                      class="visually-hidden"
                      type="color"
                      :value="getColorValue(field)"
                      @input="updateField(field.key, $event.target.value)"
                    >
                    <span
                      class="visual-template-color-swatch"
                      :style="{ backgroundColor: getColorValue(field) }"
                      aria-hidden="true"
                    ></span>
                    <span>Personnalisée</span>
                  </label>
                </div>

                <div v-else-if="field.type === 'threshold-colors'" class="d-flex flex-column gap-2">
	                  <div
	                    v-for="(rule, ruleIndex) in thresholdRulesValue(field)"
	                    :key="ruleIndex"
	                    class="visual-template-rule-card border rounded-3 p-2 bg-body-tertiary"
	                    :class="ruleDragClass(field, ruleIndex)"
	                    draggable="true"
	                    @dragstart="onRuleDragStart(field, ruleIndex, $event)"
	                    @dragover="onRuleDragOver(field, ruleIndex, $event)"
	                    @dragleave="onRuleDragLeave(field, ruleIndex)"
	                    @dragend="clearRuleDragState"
	                    @drop="onRuleDrop(field, ruleIndex)"
	                  >
	                    <div class="row g-2 align-items-center">
	                      <div class="col-auto">
	                        <div class="visual-template-rule-drag-handle">
	                          <i class="bi bi-grip-vertical"></i>
	                        </div>
	                      </div>
	                      <div class="col-12 col-md-2">
                        <select
                          class="form-select form-select-sm"
                          :value="rule.operator || 'lt'"
                          @change="updateThresholdRule(field, ruleIndex, 'operator', $event.target.value)"
                        >
                          <option value="lt">Inférieur à</option>
                          <option value="lte">Inférieur ou égal à</option>
                          <option value="gt">Supérieur à</option>
                          <option value="gte">Supérieur ou égal à</option>
                          <option value="between">Entre</option>
                        </select>
                      </div>
                      <div class="col-6 col-md-2">
                        <input
                          class="form-control form-control-sm"
                          type="number"
                          step="0.01"
                          :value="rule.value"
                          placeholder="0"
                          @input="updateThresholdRule(field, ruleIndex, 'value', $event.target.value)"
                        >
                      </div>
                      <div v-if="rule.operator === 'between'" class="col-6 col-md-2">
                        <input
                          class="form-control form-control-sm"
                          type="number"
                          step="0.01"
                          :value="rule.valueMax"
                          placeholder="100"
                          @input="updateThresholdRule(field, ruleIndex, 'valueMax', $event.target.value)"
                        >
                      </div>
                      <div class="col-8 col-md-2">
                        <select
                          class="form-select form-select-sm"
                          :value="getRuleAccentSelectValue(rule) || 'info'"
                          @change="updateRuleAccent(field, ruleIndex, $event.target.value, '#0dcaf0')"
                        >
                          <option v-for="option in field.options" :key="optionValue(option)" :value="optionValue(option)">
                            {{ optionLabel(option) }}
                          </option>
                          <option value="custom">Personnalisée</option>
                        </select>
                      </div>
                      <div v-if="isHtmlColor(rule.accent)" class="col-4 col-md-auto">
                        <label class="input-group input-group-sm visual-template-rule-color">
                          <input
                            class="form-control form-control-color"
                            type="color"
                            :value="getRuleColorValue(rule, '#0dcaf0')"
                            title="Couleur personnalisée"
                            @input="updateThresholdRule(field, ruleIndex, 'accent', $event.target.value)"
                          >
                        </label>
                      </div>
                      <div class="col-3 col-md-auto ms-md-auto text-end">
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-danger"
                          title="Supprimer"
                          @click="removeThresholdRule(field, ruleIndex)"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn btn-sm btn-outline-primary align-self-start d-inline-flex align-items-center gap-1"
                    @click="addThresholdRule(field)"
                  >
                    <i class="bi bi-plus-lg"></i>
                    <span>Ajouter une règle</span>
                  </button>
                </div>

                <div v-else-if="field.type === 'condition-colors'" class="d-flex flex-column gap-2">
	                  <div
	                    v-for="(rule, ruleIndex) in conditionRulesValue(field)"
	                    :key="ruleIndex"
	                    class="visual-template-rule-card border rounded-3 p-2 bg-body-tertiary"
	                    :class="ruleDragClass(field, ruleIndex)"
	                    draggable="true"
	                    @dragstart="onRuleDragStart(field, ruleIndex, $event)"
	                    @dragover="onRuleDragOver(field, ruleIndex, $event)"
	                    @dragleave="onRuleDragLeave(field, ruleIndex)"
	                    @dragend="clearRuleDragState"
	                    @drop="onRuleDrop(field, ruleIndex)"
	                  >
	                    <div class="row g-2 align-items-center">
	                      <div class="col-auto">
	                        <div class="visual-template-rule-drag-handle">
	                          <i class="bi bi-grip-vertical"></i>
	                        </div>
	                      </div>
	                      <div class="col-12 col-md-2">
                        <div class="input-group input-group-sm">
                          <input
                            class="form-control"
                            type="text"
                            :value="rule.field"
                            :placeholder="templateMeta.config?.badgeField || 'status'"
                            @input="updateConditionRule(field, ruleIndex, 'field', $event.target.value)"
                          >
                          <button
                            type="button"
                            class="btn btn-outline-secondary"
                            title="Rechercher une clé"
                            @click="toggleConditionKeyPicker(field, ruleIndex)"
                          >
                            <i class="bi bi-search"></i>
                          </button>
                        </div>
                      </div>
                      <div class="col-12 col-md-2">
                        <select
                          class="form-select form-select-sm"
                          :value="rule.operator || 'contains'"
                          @change="updateConditionRule(field, ruleIndex, 'operator', $event.target.value)"
                        >
                          <option value="eq">Égal à</option>
                          <option value="neq">Différent de</option>
                          <option value="contains">Contient</option>
                          <option value="not_contains">Ne contient pas</option>
                          <option value="lt">Inférieur à</option>
                          <option value="lte">Inférieur ou égal à</option>
                          <option value="gt">Supérieur à</option>
                          <option value="gte">Supérieur ou égal à</option>
                          <option value="between">Entre</option>
                          <option value="empty">Est vide</option>
                          <option value="not_empty">N’est pas vide</option>
                        </select>
                      </div>
                      <div v-if="!['empty', 'not_empty'].includes(rule.operator)" class="col-6 col-md-2">
                        <input
                          class="form-control form-control-sm"
                          type="text"
                          :inputmode="['lt', 'lte', 'gt', 'gte', 'between'].includes(rule.operator) ? 'decimal' : 'text'"
                          :value="rule.value"
                          placeholder="OK ou {{ item.champ }}"
                          @input="updateConditionRule(field, ruleIndex, 'value', $event.target.value)"
                        >
                      </div>
                      <div v-if="rule.operator === 'between'" class="col-6 col-md-2">
                        <input
                          class="form-control form-control-sm"
                          type="text"
                          inputmode="decimal"
                          :value="rule.valueMax"
                          placeholder="100 ou {{ item.max }}"
                          @input="updateConditionRule(field, ruleIndex, 'valueMax', $event.target.value)"
                        >
                      </div>
                      <div class="col-8 col-md-2">
                        <select
                          class="form-select form-select-sm"
                          :value="getRuleAccentSelectValue(rule) || 'success'"
                          @change="updateRuleAccent(field, ruleIndex, $event.target.value, '#198754')"
                        >
                          <option v-for="option in field.options" :key="optionValue(option)" :value="optionValue(option)">
                            {{ optionLabel(option) }}
                          </option>
                          <option value="custom">Personnalisée</option>
                        </select>
                      </div>
                      <div v-if="isHtmlColor(rule.accent)" class="col-4 col-md-auto">
                        <label class="input-group input-group-sm visual-template-rule-color">
                          <input
                            class="form-control form-control-color"
                            type="color"
                            :value="getRuleColorValue(rule, '#198754')"
                            title="Couleur personnalisée"
                            @input="updateConditionRule(field, ruleIndex, 'accent', $event.target.value)"
                          >
                        </label>
                      </div>
                      <div class="col-3 col-md-auto ms-md-auto text-end">
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-danger"
                          title="Supprimer"
                          @click="removeConditionRule(field, ruleIndex)"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                      <div v-if="conditionKeyPickerOpen === `${field.key}:${ruleIndex}`" class="col-12">
                        <div class="border rounded-3 bg-body p-2">
                          <div v-if="conditionKeysError" class="text-warning small">
                            <i class="bi bi-exclamation-triangle-fill me-1"></i>{{ conditionKeysError }}
                          </div>
                          <div v-else-if="conditionKeysLoading" class="d-flex align-items-center gap-2 text-secondary small">
                            <span class="spinner-border spinner-border-sm" role="status"></span>
                            <span>Chargement...</span>
                          </div>
                          <div v-else-if="conditionKeys.length" class="d-flex flex-wrap gap-2">
                            <button
                              v-for="keyName in conditionKeys"
                              :key="keyName"
                              type="button"
                              class="btn btn-sm btn-outline-primary font-monospace"
                              @click="selectConditionKey(field, ruleIndex, keyName)"
                            >
                              {{ keyName }}
                            </button>
                          </div>
                          <div v-else class="text-secondary small">
                            Aucune clé détectée dans le premier élément.
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn btn-sm btn-outline-primary align-self-start d-inline-flex align-items-center gap-1"
                    @click="addConditionRule(field)"
                  >
                    <i class="bi bi-plus-lg"></i>
                    <span>Ajouter une règle</span>
                  </button>
                </div>
                
                <div v-else-if="field.type === 'template'">
                  <NunjucksTemplateEditor
                    :id="`visual-${field.key}`"
                    :model-value="getFieldDisplayValue(field)"
                    :placeholder="fieldPlaceholder(field)"
                    min-height="7rem"
                    @update:model-value="updateField(field.key, $event)"
                  />
                  <div v-if="selectedTemplateHasCollection" class="d-flex flex-wrap gap-2 mt-2">
                    <button type="button" class="btn btn-xs btn-outline-secondary font-monospace" @click="appendTemplateToken(field, '{{ value }}')" v-text="'{{ value }}'"></button>
                    <button type="button" class="btn btn-xs btn-outline-secondary font-monospace" @click="appendTemplateToken(field, '{{ item }}')" v-text="'{{ item }}'"></button>
                    <button type="button" class="btn btn-xs btn-outline-secondary font-monospace" @click="appendTemplateToken(field, '{{ item.total }}')" v-text="'{{ item.total }}'"></button>
                  </div>
                </div>

                <div v-else-if="field.type === 'template-variables'" class="d-flex flex-column gap-2">
                  <div
                    v-for="(variable, variableIndex) in templateVariablesValue(field)"
                    :key="variableIndex"
                    class="visual-template-rule-card border rounded-3 p-2 bg-body-tertiary"
                  >
                    <div class="row g-2 align-items-start">
                      <div class="col-12 col-md-3">
                        <input
                          class="form-control form-control-sm font-monospace"
                          :class="{ 'is-invalid': variable.name && !validTemplateVariableName(variable.name) }"
                          type="text"
                          :value="variable.name"
                          placeholder="nom"
                          @input="updateTemplateVariable(field, variableIndex, 'name', $event.target.value)"
                        >
                        <div v-if="variable.name && !validTemplateVariableName(variable.name)" class="invalid-feedback">
                          Lettres, chiffres et _ uniquement. Le nom doit commencer par une lettre ou _.
                        </div>
                        <div
                          v-else-if="templateVariableToken(field, variable)"
                          class="input-group input-group-sm mt-2"
                        >
                          <code
                            class="form-control bg-body font-monospace small text-truncate"
                            v-text="templateVariableToken(field, variable)"
                          ></code>
                          <button
                            type="button"
                            class="btn btn-outline-secondary"
                            :title="copiedTemplateVariableKey === `${field.key}:${variableIndex}` ? 'Copié' : 'Copier'"
                            @click="copyTemplateVariableToken(`${field.key}:${variableIndex}`, templateVariableToken(field, variable))"
                          >
                            <i :class="copiedTemplateVariableKey === `${field.key}:${variableIndex}` ? 'bi bi-check2' : 'bi bi-clipboard'"></i>
                          </button>
                        </div>
                      </div>
                      <div class="col-12 col-md">
                        <NunjucksTemplateEditor
                          :model-value="variable.value"
                          :placeholder="fieldPlaceholder(field)"
                          min-height="5.5rem"
                          @update:model-value="updateTemplateVariable(field, variableIndex, 'value', $event)"
                        />
                      </div>
                      <div class="col-12 col-md-auto text-end">
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-danger"
                          title="Supprimer"
                          @click="removeTemplateVariable(field, variableIndex)"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="btn btn-sm btn-outline-primary align-self-start d-inline-flex align-items-center gap-1"
                    @click="addTemplateVariable(field)"
                  >
                    <i class="bi bi-plus-lg"></i>
                    <span>Ajouter une variable</span>
                  </button>
                </div>

                <div v-else-if="field.type === 'range'" class="d-flex flex-column gap-2">
                  <div class="d-flex align-items-center gap-3">
                    <input
                      :id="`visual-${field.key}`"
                      class="form-range flex-grow-1"
                      type="range"
                      :min="field.min ?? 0"
                      :max="field.max ?? 1"
                      :step="field.step ?? 0.05"
                      :value="getRangeFieldValue(field)"
                      @input="updateField(field.key, $event.target.value)"
                    >
                    <span class="badge text-bg-secondary font-monospace">{{ getRangeFieldValue(field).toFixed(2) }}</span>
                  </div>
                  <div class="d-flex justify-content-between text-secondary small">
                    <span>{{ field.min ?? 0 }}</span>
                    <span>{{ field.max ?? 1 }}</span>
                  </div>
                </div>
                
                <DataPathPicker
                  v-else-if="field.type === 'data-path'"
                  :field="field"
                  :template-meta="templateMeta"
                  :item-id="itemId"
                  :source-list-version="sourceListVersion"
                  @update:template-meta="emit('update:templateMeta', $event)"
                  @update-field="updateField"
                />
                
                <DataFieldPicker
                  v-else-if="isObjectKeyField(field)"
                  :field="field"
                  :template-meta="templateMeta"
                  :item-id="itemId"
                  @update:template-meta="emit('update:templateMeta', $event)"
                  @update-field="updateField"
                />

                <DataValuePicker
                  v-else-if="isAbsoluteValueField(field)"
                  :field="field"
                  :template-meta="templateMeta"
                  :item-id="itemId"
                  @update:template-meta="emit('update:templateMeta', $event)"
                  @update-field="updateField"
                />

                <DataTableColumnsManager
                  v-else-if="field.type === 'columns-manager'"
                  :field="field"
                  :template-meta="templateMeta"
                  :item-id="itemId"
                  @update:template-meta="emit('update:templateMeta', $event)"
                  @update-field="updateField"
                />
                
                <input v-else :id="`visual-${field.key}`" class="form-control shadow-xs" :type="field.type"
                  :value="getFieldDisplayValue(field)" :required="field.required"
                  :placeholder="fieldPlaceholder(field)"
                  @input="updateField(field.key, $event.target.value)">
              </div>
              <div v-if="fieldIssue(field)" class="form-text mt-1 text-warning" style="font-size: 0.78rem;">{{ fieldIssue(field).message }}</div>
              <div v-else-if="field.help && field.type !== 'columns-manager'" class="form-text mt-1 text-secondary opacity-75" style="font-size: 0.78rem;">{{ field.help }}</div>
            </div>
          </div>
              </div>
              </div>
            </div>
          </section>
        </div>

      </div>
      
      <div v-else class="alert alert-warning mb-0" role="alert">
        {{ $t('edititem.templates.unavailable') }}
      </div>
    </div>
  </article>

</template>

<style scoped>
.visual-template-workspace {
  max-width: 100%;
}

.visual-template-section {
  overflow: hidden;
  border: 1px solid var(--dc-border-subtle) !important;
  border-radius: 12px;
  background: var(--bs-body-bg);
  transition: var(--dc-transition-smooth);
  position: relative;
}

/* Sections color themes variables */
.section-data {
  --section-accent-color: #3b82f6;
  --section-accent-bg: rgba(59, 130, 246, 0.08);
}
.section-display {
  --section-accent-color: #8b5cf6;
  --section-accent-bg: rgba(139, 92, 246, 0.08);
}
.section-colors {
  --section-accent-color: #ec4899;
  --section-accent-bg: rgba(236, 72, 153, 0.08);
}
.section-options {
  --section-accent-color: #10b981;
  --section-accent-bg: rgba(16, 185, 129, 0.08);
}
.section-advanced {
  --section-accent-color: #f97316;
  --section-accent-bg: rgba(249, 115, 22, 0.08);
}
.section-variables {
  --section-accent-color: #eab308;
  --section-accent-bg: rgba(234, 179, 8, 0.08);
}

/* Left vertical indicator */
.visual-template-section::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: transparent;
  transition: var(--dc-transition-fast);
  z-index: 10;
}

.visual-template-section.is-open::before {
  background-color: var(--section-accent-color);
}

.visual-template-section:hover, .visual-template-section.is-open {
  border-color: var(--section-accent-color) !important;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.03);
}

.visual-template-section-toggle {
  cursor: pointer;
  transition: var(--dc-transition-fast);
}

.visual-template-section-toggle:hover {
  background-color: var(--bs-tertiary-bg) !important;
}

.visual-template-section-icon {
  width: 2.5rem;
  height: 2.5rem;
  color: var(--section-accent-color, var(--bs-primary));
  background: var(--section-accent-bg, rgba(var(--bs-primary-rgb), 0.08));
  border-radius: 10px;
  flex: 0 0 auto;
  font-size: 1.1rem;
  transition: var(--dc-transition-fast);
}

.visual-template-section-toggle:hover .visual-template-section-icon {
  background: var(--section-accent-color) !important;
  color: #fff !important;
  transform: scale(1.05);
}

.visual-template-field-group + .visual-template-field-group {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--dc-border-subtle);
  padding-top: 1.5rem;
}

.visual-template-field-group-title {
  margin-bottom: 1rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--bs-secondary-color);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.visual-template-field-group-title::after {
  content: '';
  flex: 1 1 auto;
  height: 1px;
  background: var(--dc-border-subtle);
}

/* Rule cards styling (threshold and conditions) */
:deep(.visual-template-rule-card) {
  border: 1px solid var(--dc-border-subtle) !important;
  background: rgba(var(--bs-tertiary-bg-rgb), 0.3) !important;
  border-radius: 10px !important;
  transition: var(--dc-transition-smooth);
}

:deep(.visual-template-rule-card:hover) {
  border-color: var(--dc-hover-border) !important;
  background: rgba(var(--bs-tertiary-bg-rgb), 0.5) !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

:deep(.visual-template-rule-drag-handle) {
  cursor: grab;
  color: var(--bs-secondary-color);
  opacity: 0.5;
  transition: var(--dc-transition-fast);
  padding: 0.25rem;
}

:deep(.visual-template-rule-drag-handle:hover) {
  opacity: 1;
  color: var(--bs-primary);
}

:deep(.visual-template-rule-card.dragging) {
  opacity: 0.4;
  border-style: dashed !important;
}

/* Custom inputs styling inside visual editor */
:deep(.form-control),
:deep(.form-select) {
  border-radius: 8px;
  border-color: var(--dc-border-subtle);
  transition: var(--dc-transition-fast);
}

:deep(.form-control:focus),
:deep(.form-select:focus) {
  border-color: rgba(59, 130, 246, 0.55) !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15) !important;
}
</style>
