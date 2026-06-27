<script setup>
import { computed, inject, ref } from 'vue';
import axios from 'axios';
import { getTemplateDefinition } from '@/templates/catalog.js';
import DataPathPicker from './DataPathPicker.vue';
import DataFieldPicker from './DataFieldPicker.vue';
import DataValuePicker from './DataValuePicker.vue';
import DataTableColumnsManager from './DataTableColumnsManager.vue';

const props = defineProps({
  templateMeta: { type: Object, required: true },
  itemId: { type: [String, Number], required: true }
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

function isObjectKeyField(field) {
  return field.key.endsWith('Field');
}

function isAbsoluteValueField(field) {
  return field.key === 'value' || field.key === 'prevValue';
}

const selectedTemplate = computed(() => getTemplateDefinition(
  props.templateMeta.templateKey,
  props.templateMeta.templateMajor
) || null);

function updateField(key, value) {
  emit('update:templateMeta', {
    ...props.templateMeta,
    config: { ...props.templateMeta.config, [key]: value }
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

const fields = computed(() => selectedTemplate.value?.fields || []);
const booleanFields = computed(() => fields.value.filter(isBooleanField));
const regularFields = computed(() => fields.value.filter(f => !isBooleanField(f)));
const templateItemsPath = computed(() => props.templateMeta.config?.items ?? '');

function getFieldClass(field) {
  if (field.type === 'columns-manager') return 'col-12';
  if (field.type === 'condition-colors') return 'col-12';
  if (field.type === 'threshold-colors') return 'col-12';
  if (field.type === 'template') return 'col-12';
  if (field.key === 'title' || field.key === 'items') return 'col-12 col-md-6';
  if (field.type === 'number') return 'col-6 col-md-4';
  if (field.type === 'select') return 'col-12 col-md-6';
  if (field.type === 'color') return 'col-12';
  return 'col-12 col-md-6';
}

function isFieldValueTrue(field) {
  const val = props.templateMeta.config?.[field.key];
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

    const savedExamples = props.templateMeta.sourceExamples?.[source.id] || {};
    const res = await axios.get(`${apiUrl}/data/source/${source.id}`, {
      params: savedExamples
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
      <div v-if="selectedTemplate" class="d-flex flex-column gap-4">
        <!-- Champs principaux -->
        <div class="row g-3">
          <div 
            v-for="field in regularFields" 
            :key="field.key" 
            :class="getFieldClass(field)"
          >
            <div class="visual-template-field h-100 d-flex flex-column justify-content-between">
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
                
                <textarea v-else-if="field.type === 'template'" :id="`visual-${field.key}`" class="form-control font-monospace shadow-xs"
                  rows="3" :value="getFieldDisplayValue(field)" :required="field.required"
                  :placeholder="fieldPlaceholder(field)"
                  @input="updateField(field.key, $event.target.value)"></textarea>
                
                <DataPathPicker
                  v-else-if="field.type === 'data-path'"
                  :field="field"
                  :template-meta="templateMeta"
                  :item-id="itemId"
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
              <div v-if="field.help && field.type !== 'columns-manager'" class="form-text mt-1 text-secondary opacity-75" style="font-size: 0.78rem;">{{ field.help }}</div>
            </div>
          </div>
        </div>

        <!-- Section Options / Fonctionnalités -->
        <div v-if="booleanFields.length > 0" class="pt-3 border-top">
          <h5 class="h6 mb-3 text-secondary d-flex align-items-center gap-2">
            <i class="bi bi-toggle-on"></i>
            <span>{{ $t('edititem.visual.options_section') }}</span>
          </h5>
          
          <div class="row g-3">
            <div 
              v-for="field in booleanFields" 
              :key="field.key" 
              class="col-12 col-md-6 col-lg-4"
            >
              <div 
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
            </div>
          </div>
        </div>

      </div>
      
      <div v-else class="alert alert-warning mb-0" role="alert">
        {{ $t('edititem.templates.unavailable') }}
      </div>
    </div>
  </article>

</template>
