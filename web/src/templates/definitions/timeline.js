import {
  accentCssColor,
  accentOptions,
  collectionTemplate,
  compileTemplateVariables,
  createTemplateDefinition,
  dataPath,
  itemOutput,
  normalizeAccent,
  normalizeTemplateVariables,
  templateVariableFields
} from '../helpers.js';

export const timeline = createTemplateDefinition({
  key: 'template/timeline',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  category: 'Listes',
  useCase: 'Présenter des événements dans l’ordre chronologique.',
  name: 'Timeline',
  description: 'Affiche les événements récents sous forme de fil chronologique.',
  preview: '<div class="border-start ps-3"><div class="mb-3"><div class="small text-secondary">09:30</div><div class="fw-medium">Import terminé</div><div class="small text-secondary">1 248 lignes traitées</div></div><div><div class="small text-secondary">10:15</div><div class="fw-medium">Rapport publié</div><div class="small text-secondary">Finance</div></div></div>',
  fields: [
    { key: 'title', type: 'text', section: 'display', required: true, placeholder: 'Activité', help: 'Titre du fil chronologique.' },
    { key: 'items', type: 'data-path', section: 'data', required: true, placeholder: 'events', help: 'Nom de la liste d’événements dans les données.' },
    { key: 'dateField', type: 'text', section: 'data', required: true, placeholder: 'date', help: 'Champ contenant la date ou l’heure.' },
    { key: 'labelField', type: 'text', section: 'data', required: true, placeholder: 'title', help: 'Champ contenant le titre de l’événement.' },
    { key: 'detailField', type: 'text', section: 'data', placeholder: 'detail', help: 'Champ optionnel affiché sous l’événement.' },
    { key: 'accent', type: 'color', section: 'colors', options: accentOptions },
    ...templateVariableFields
  ],
  defaults: {
    title: 'Activité',
    items: '',
    dateField: '',
    labelField: '',
    detailField: '',
    accent: 'primary',
    globalVariables: '',
    loopVariables: ''
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      items: dataPath(config.items, ''),
      dateField: dataPath(config.dateField, ''),
      labelField: dataPath(config.labelField, ''),
      detailField: String(config.detailField ?? '').trim(),
      accent: normalizeAccent(config.accent, this.defaults.accent),
      globalVariables: normalizeTemplateVariables(config.globalVariables),
      loopVariables: normalizeTemplateVariables(config.loopVariables)
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.items) {
      return {
        template: '<div class="text-secondary small">Veuillez configurer la source de données (items)</div>',
        javascript: ''
      };
    }
    const detail = normalized.detailField
      ? `<div class="small text-secondary">${itemOutput(normalized.detailField, "''")}</div>`
      : '';
    const itemHtml = `<div class="mb-3 position-relative"><div class="small text-secondary">${itemOutput(normalized.dateField, 'key')}</div><div class="fw-medium">${itemOutput(normalized.labelField, 'value')}</div>${detail}</div>`;
    const emptyHtml = '<div class="text-secondary">Aucune donnée</div>';

    const globalSetup = compileTemplateVariables(normalized.globalVariables);

    return {
      template: `${globalSetup}<div class="border-start ps-3" style="border-left-color: ${accentCssColor(normalized.accent)} !important;">${collectionTemplate(normalized.items, itemHtml, itemHtml, emptyHtml, normalized.loopVariables, normalized.globalVariables)}</div>`,
      javascript: ''
    };
  }
});
