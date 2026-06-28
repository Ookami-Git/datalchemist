import {
  booleanOption,
  booleanOptions,
  collectionTemplate,
  compileTemplateVariables,
  createTemplateDefinition,
  dataPath,
  dataTableButtonOptions,
  hashString,
  integerOption,
  itemExpression,
  orderDirectionOptions,
  normalizeTemplateVariables,
  templateVariableFields,
  toBoolean
} from '../helpers.js';

export const compactTable = createTemplateDefinition({
  key: 'template/compact-table',
  major: 1,
  configVersion: 1,
  helpSections: ['nunjucks'],
  category: 'Tableaux',
  useCase: 'Explorer une collection avec tri, recherche, pagination et exports.',
  name: 'Tableau DataTables',
  description: 'Génère les colonnes depuis une collection et active recherche, tri, pagination, export et options avancées DataTables.',
  preview: '<table class="table table-sm table-striped align-middle w-100"><thead><tr><th>Client</th><th>Segment</th><th>CA</th></tr></thead><tbody><tr><td>Acme</td><td>Enterprise</td><td>18 k€</td></tr><tr><td>Northwind</td><td>Retail</td><td>12 k€</td></tr></tbody></table>',
  fields: [
    { key: 'title', type: 'text', section: 'display', required: true, placeholder: 'Top clients', help: 'Titre du tableau.' },
    { key: 'items', type: 'data-path', section: 'data', required: true, placeholder: 'clients', help: 'Nom de la liste dans les données. Le tableau crée les colonnes automatiquement avec les clés du premier objet.' },
    { key: 'columns', type: 'columns-manager', section: 'data', help: 'Personnalisez et ordonnez les colonnes du tableau, ou laissez vide pour générer automatiquement toutes les colonnes.' },
    { key: 'pageLength', type: 'number', section: 'display', placeholder: '25', help: 'Nombre de lignes affichées par page.' },
    { key: 'searching', type: 'select', section: 'options', options: booleanOptions, help: 'Active le champ de recherche DataTables.' },
    { key: 'ordering', type: 'select', section: 'options', options: booleanOptions, help: 'Active le tri sur les colonnes.' },
    { key: 'paging', type: 'select', section: 'options', options: booleanOptions, help: 'Active la pagination.' },
    { key: 'info', type: 'select', section: 'options', options: booleanOptions, help: 'Affiche les informations de pagination.' },
    { key: 'responsive', type: 'select', section: 'options', options: booleanOptions, help: 'Adapte les colonnes aux petits écrans.' },
    { key: 'buttons', type: 'select', section: 'advanced', options: dataTableButtonOptions, help: 'Ajoute les boutons de copie, export, impression et visibilité des colonnes.' },
    { key: 'searchBuilder', type: 'select', section: 'advanced', options: booleanOptions, help: 'Ajoute le constructeur de recherche avancée.' },
    { key: 'searchPanes', type: 'select', section: 'advanced', options: booleanOptions, help: 'Ajoute les panneaux de filtrage par colonne.' },
    { key: 'colReorder', type: 'select', section: 'advanced', options: booleanOptions, help: 'Permet de réordonner les colonnes.' },
    { key: 'fixedHeader', type: 'select', section: 'advanced', options: booleanOptions, help: 'Garde l’en-tête visible pendant le scroll.' },
    { key: 'stateSave', type: 'select', section: 'advanced', options: booleanOptions, help: 'Mémorise recherche, tri, page et colonnes dans le navigateur.' },
    { key: 'selectRows', type: 'select', section: 'advanced', options: booleanOptions, help: 'Permet la sélection de lignes.' },
    { key: 'scrollX', type: 'select', section: 'advanced', options: booleanOptions, help: 'Active le défilement horizontal.' },
    { key: 'scroller', type: 'select', section: 'advanced', options: booleanOptions, help: 'Optimise l’affichage des grandes listes.' },
    { key: 'fixedColumnsLeft', type: 'number', section: 'advanced', placeholder: '0', help: 'Nombre de colonnes figées à gauche.' },
    { key: 'fixedColumnsRight', type: 'number', section: 'advanced', placeholder: '0', help: 'Nombre de colonnes figées à droite.' },
    { key: 'rowGroupColumn', type: 'number', section: 'advanced', placeholder: '0', help: 'Index de colonne à regrouper, en partant de 1. 0 désactive le regroupement.' },
    { key: 'orderColumn', type: 'number', section: 'advanced', placeholder: '1', help: 'Index de colonne utilisée pour le tri initial, en partant de 1.' },
    { key: 'orderDirection', type: 'select', section: 'advanced', options: orderDirectionOptions, help: 'Sens du tri initial.' },
    ...templateVariableFields
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
    orderDirection: 'asc',
    globalVariables: '',
    loopVariables: ''
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
      orderDirection,
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
    const globalSetup = compileTemplateVariables(normalized.globalVariables);
    const loopSetup = compileTemplateVariables(normalized.loopVariables, normalized.globalVariables, 'loop');

    let templateHtml = '';
    let parsedColumns = [];
    if (normalized.columns) {
      try {
        parsedColumns = JSON.parse(normalized.columns);
      } catch {
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
      templateHtml = `{% if ${normalized.items} %}<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}"><thead><tr>${theadRows}</tr></thead><tbody>${collectionTemplate(normalized.items, rowHtml, rowHtml, `<tr><td colspan="${parsedColumns.length}">Aucune donnée</td></tr>`, normalized.loopVariables, normalized.globalVariables)}</tbody></table>{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    } else {
      const arrayTable = `<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}"><thead><tr>{% for key, value in ${normalized.items}[0] %}<th>{{ key }}</th>{% endfor %}</tr></thead><tbody>{% for item in ${normalized.items} %}{% set key = loop.index0 %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}<tr>{% for key, value in item %}<td>{{ value }}</td>{% endfor %}</tr>{% endfor %}</tbody></table>`;
      const objectTable = `<table class="table table-sm table-striped table-hover align-middle w-100" data-dc-datatable="${tableToken}">{% for key, item in ${normalized.items} %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}{% if loop.first %}<thead><tr><th>Clé</th>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<th>{{ fieldKey }}</th>{% endfor %}{% else %}<th>Valeur</th>{% endif %}</tr></thead><tbody>{% endif %}<tr><td>{{ key }}</td>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<td>{{ fieldValue }}</td>{% endfor %}{% else %}<td>{{ value }}</td>{% endif %}</tr>{% else %}<tbody><tr><td>Aucune donnée</td></tr>{% endfor %}</tbody></table>`;
      templateHtml = `{% if ${normalized.items} and ${normalized.items}.length %}${arrayTable}{% elif ${normalized.items} %}${objectTable}{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    }

    return {
      template: `${globalSetup}${templateHtml}`,
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
});
