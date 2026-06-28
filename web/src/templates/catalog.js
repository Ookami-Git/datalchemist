import { progressCard } from './definitions/progressCard.js';
import { statusList } from './definitions/statusList.js';
import { compactTable } from './definitions/compactTable.js';
import { timeline } from './definitions/timeline.js';
import { formatTemplateForHumans } from './helpers.js';

export const templateCatalog = [
  progressCard,
  statusList,
  compactTable,
  timeline
];

export function getTemplateDefinition(key, major) {
  return templateCatalog.find((template) => template.key === key && template.major === Number(major)) || null;
}

export function migrateTemplateConfig(definition, config = {}, fromVersion = definition?.configVersion) {
  if (!definition) return config;
  const migrated = typeof definition.migrateConfig === 'function'
    ? definition.migrateConfig(config, fromVersion)
    : { ...config };
  return definition.normalizeConfig(migrated);
}

export function validateTemplateConfig(definition, config = {}) {
  if (!definition || typeof definition.validateConfig !== 'function') return [];
  return definition.validateConfig(config);
}

export function compileTemplateDefinition(definition, config = {}, context = {}) {
  if (!definition) {
    return {
      template: '<div class="alert alert-danger mb-0">Visual template unavailable.</div>',
      javascript: ''
    };
  }
  const compiled = definition.compile(migrateTemplateConfig(definition, config), context);
  return {
    ...compiled,
    template: formatTemplateForHumans(compiled.template)
  };
}
