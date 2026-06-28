import {
  accentOptions,
  compileProgressAccentRules,
  compileTemplateVariables,
  createTemplateDefinition,
  escapeHtml,
  normalizeAccent,
  normalizeProgressColorRules,
  normalizeTemplateVariables,
  templateExpression,
  templateVariableFields
} from '../helpers.js';

export const progressCard = createTemplateDefinition({
  key: 'template/progress-card',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  category: 'Indicateurs',
  useCase: 'Suivre une valeur par rapport à un objectif.',
  name: 'Progression',
  description: 'Affiche un objectif, sa valeur actuelle et une jauge de progression.',
  preview: '<div class="d-flex justify-content-between"><div><div class="fw-semibold">Objectif mensuel</div><div class="small text-secondary">75 / 100 dossiers</div></div><div class="fw-semibold">75%</div></div><div class="progress mt-3" role="progressbar" aria-valuenow="75" aria-valuemin="0" aria-valuemax="100"><div class="progress-bar bg-info" style="width: 75%"></div></div>',
  fields: [
    { key: 'title', type: 'text', section: 'data', required: true, placeholder: 'Objectif mensuel', help: 'Nom de l’objectif affiché.' },
    { key: 'current', type: 'template', section: 'data', required: true, placeholder: '{{ done }}', help: 'Valeur dynamique actuelle.' },
    { key: 'target', type: 'template', section: 'data', required: true, placeholder: '{{ target }}', help: 'Valeur dynamique cible.' },
    { key: 'percent', type: 'template', section: 'data', required: true, placeholder: '{{ percent }}', help: 'Pourcentage dynamique entre 0 et 100.' },
    { key: 'unit', type: 'text', section: 'display', placeholder: 'dossiers', help: 'Unité optionnelle affichée après la cible.' },
    { key: 'accent', type: 'color', section: 'colors', options: accentOptions, help: 'Couleur utilisée si aucune règle de seuil ne correspond.' },
    { key: 'colorRules', type: 'threshold-colors', section: 'colors', options: accentOptions, help: 'Règles de couleur évaluées dans l’ordre sur le pourcentage.' },
    templateVariableFields[0]
  ],
  defaults: {
    title: 'Objectif',
    current: '',
    target: '',
    percent: '',
    unit: '',
    accent: 'info',
    colorRules: [],
    globalVariables: ''
  },
  normalizeConfig(config = {}) {
    return {
      title: String(config.title ?? this.defaults.title),
      current: templateExpression(config.current, ''),
      target: templateExpression(config.target, ''),
      percent: templateExpression(config.percent, ''),
      unit: String(config.unit ?? this.defaults.unit),
      accent: normalizeAccent(config.accent, this.defaults.accent),
      colorRules: normalizeProgressColorRules(config.colorRules),
      globalVariables: normalizeTemplateVariables(config.globalVariables)
    };
  },
  compile(config) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.current || !normalized.target) {
      return {
        template: `<div class="d-flex justify-content-between gap-3"><div><div class="fw-semibold">${escapeHtml(normalized.title)}</div><div class="small text-secondary">Non configuré</div></div></div>`,
        javascript: ''
      };
    }
    const unit = normalized.unit ? ` ${escapeHtml(normalized.unit)}` : '';
    const progressColor = compileProgressAccentRules(normalized.colorRules, normalized.accent);
    const globalSetup = compileTemplateVariables(normalized.globalVariables);
    const progressPercentSetup = normalized.colorRules.length
      ? `{% set progressPercentValue %}${normalized.percent || 0}{% endset %}{% set progressPercent = progressPercentValue | float %}`
      : '';

    return {
      template: `${globalSetup}${progressPercentSetup}<div class="d-flex justify-content-between gap-3"><div><div class="fw-semibold">${escapeHtml(normalized.title)}</div><div class="small text-secondary">${normalized.current} / ${normalized.target}${unit}</div></div><div class="fw-semibold">${normalized.percent || 0}%</div></div><div class="progress mt-3" role="progressbar" aria-valuenow="${normalized.percent || 0}" aria-valuemin="0" aria-valuemax="100"><div class="progress-bar" style="width: ${normalized.percent || 0}%; background-color: ${progressColor};"></div></div>`,
      javascript: ''
    };
  }
});
