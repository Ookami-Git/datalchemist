import {
  accentOptions,
  collectionTemplate,
  compileStatusAccentRules,
  compileTemplateVariables,
  createTemplateDefinition,
  dataPath,
  itemOutput,
  normalizeAccent,
  normalizeStatusColorRules,
  normalizeTemplateVariables,
  templateVariableFields
} from '../helpers.js';

export const statusList = createTemplateDefinition({
  key: 'template/status-list',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  category: 'Listes',
  useCase: 'Afficher une collection avec un badge de statut.',
  name: 'Liste de statuts',
  description: 'Liste des éléments avec libellé, détail et badge dynamique.',
  preview: '<ul class="list-group"><li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">API</div><div class="small text-secondary">Latence 42 ms</div></div><span class="badge text-bg-success align-self-center">OK</span></li><li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">Jobs</div><div class="small text-secondary">2 en attente</div></div><span class="badge text-bg-warning align-self-center">À surveiller</span></li></ul>',
  fields: [
    { key: 'title', type: 'text', section: 'display', required: true, placeholder: 'Services', help: 'Titre de la liste.' },
    { key: 'items', type: 'data-path', section: 'data', required: true, placeholder: 'services', help: 'Nom de la liste dans les données. Exemple : services, alerts, clients.' },
    { key: 'labelField', type: 'text', section: 'data', required: true, placeholder: 'name', help: 'Champ utilisé comme libellé principal pour chaque ligne.' },
    { key: 'detailField', type: 'text', section: 'data', placeholder: 'detail', help: 'Champ optionnel affiché sous le libellé.' },
    { key: 'badgeField', type: 'text', section: 'data', required: true, placeholder: 'status', help: 'Champ affiché dans le badge.' },
    { key: 'accent', type: 'color', section: 'colors', options: accentOptions, help: 'Couleur utilisée si aucune règle ne correspond.' },
    { key: 'colorRules', type: 'condition-colors', section: 'colors', options: accentOptions, help: 'Règles de couleur évaluées dans l’ordre. Le champ comparé est configurable par règle.' },
    ...templateVariableFields
  ],
  defaults: {
    title: 'Statuts',
    items: '',
    labelField: '',
    detailField: '',
    badgeField: '',
    accent: 'success',
    colorRules: [],
    globalVariables: '',
    loopVariables: ''
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      items: dataPath(config.items, ''),
      labelField: dataPath(config.labelField, ''),
      detailField: String(config.detailField ?? '').trim(),
      badgeField: dataPath(config.badgeField, ''),
      accent: normalizeAccent(config.accent, this.defaults.accent),
      colorRules: normalizeStatusColorRules(config.colorRules),
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
      ? `<div class="small text-secondary">${itemOutput(normalized.detailField, 'key')}</div>`
      : '';
    const statusColor = compileStatusAccentRules(normalized.colorRules, normalized.accent, normalized.badgeField);
    const itemHtml = `<li class="list-group-item d-flex justify-content-between gap-3"><div><div class="fw-medium">${itemOutput(normalized.labelField, 'key')}</div>${detail}</div><span class="badge align-self-center" style="background-color: ${statusColor}; color: #fff;">${itemOutput(normalized.badgeField, 'value')}</span></li>`;
    const emptyHtml = '<li class="list-group-item text-secondary">Aucune donnée</li>';

    const globalSetup = compileTemplateVariables(normalized.globalVariables);

    return {
      template: `${globalSetup}<ul class="list-group">${collectionTemplate(normalized.items, itemHtml, itemHtml, emptyHtml, normalized.loopVariables, normalized.globalVariables)}</ul>`,
      javascript: ''
    };
  }
});
