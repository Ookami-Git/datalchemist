import {
  compileTemplateDefinition,
  getTemplateDefinition,
  migrateTemplateConfig,
  templateCatalog
} from '@/templates/catalog.js';

export const FREE_ITEM_MODE = 'free';
export const VISUAL_ITEM_MODE = 'visual';

export function createVisualItemParameters(template = templateCatalog[0]) {
  if (!template) {
    return {
      mode: VISUAL_ITEM_MODE,
      templateKey: '',
      templateMajor: 1,
      configVersion: 1,
      config: {}
    };
  }

  return {
    mode: VISUAL_ITEM_MODE,
    templateKey: template.key,
    templateMajor: template.major,
    configVersion: template.configVersion,
    config: { ...template.defaults }
  };
}

export function parseItemParameters(value) {
  if (!value || typeof value !== 'string') {
    return { mode: FREE_ITEM_MODE };
  }

  try {
    const parsed = JSON.parse(value);
    if (parsed?.mode === VISUAL_ITEM_MODE && typeof parsed.templateKey === 'string') {
      const definition = getTemplateDefinition(parsed.templateKey, parsed.templateMajor);
      if (!definition) return parsed;

      return {
        ...parsed,
        configVersion: definition.configVersion,
        config: migrateTemplateConfig(definition, parsed.config, parsed.configVersion)
      };
    }
  } catch {
    // Existing free-form parameters must remain untouched.
  }

  return { mode: FREE_ITEM_MODE };
}

export function serializeVisualItemParameters(parameters) {
  const serializable = { ...(parameters || {}) };
  delete serializable.getOverrides;
  delete serializable.sourceExamples;

  return JSON.stringify(serializable);
}

export function resolveItemRenderDefinition(item) {
  const parameters = parseItemParameters(item?.parameters);
  if (parameters.mode !== VISUAL_ITEM_MODE) {
    return item;
  }

  const definition = getTemplateDefinition(parameters.templateKey, parameters.templateMajor);
  if (!definition) {
    return {
      ...item,
      template: '<div class="alert alert-danger mb-0">Visual template unavailable.</div>',
      javascript: ''
    };
  }

  const compiled = compileTemplateDefinition(definition, parameters.config);
  return { ...item, ...compiled };
}
