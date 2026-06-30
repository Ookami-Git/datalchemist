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

function numberOption(value, fallback, min = Number.NEGATIVE_INFINITY, max = Number.POSITIVE_INFINITY) {
  const parsed = Number.parseFloat(value);
  if (!Number.isFinite(parsed)) return fallback;
  return Math.min(Math.max(parsed, min), max);
}

function normalizeStringArray(value) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item ?? '').trim()).filter(Boolean);
  }

  const normalized = String(value ?? '').trim();
  if (!normalized) return [];

  try {
    return normalizeStringArray(JSON.parse(normalized));
  } catch {
    return normalized.split(',').map((item) => item.trim()).filter(Boolean);
  }
}

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
    { key: 'items', type: 'data-path', section: 'data', required: true, placeholder: 'clients', help: 'Nom de la liste dans les données. Le tableau crée les colonnes automatiquement avec les clés du premier objet.' },
    { key: 'columns', type: 'columns-manager', section: 'data', help: 'Personnalisez et ordonnez les colonnes du tableau, ou laissez vide pour générer automatiquement toutes les colonnes.' },
    { key: 'searching', type: 'select', section: 'options', group: 'search', options: booleanOptions, help: 'Active le champ de recherche DataTables.' },
    { key: 'searchBuilder', type: 'select', section: 'options', group: 'search', dependsOn: 'searching', options: booleanOptions, help: 'Ajoute le constructeur de recherche avancée.' },
    { key: 'searchPanes', type: 'select', section: 'options', group: 'search', dependsOn: 'searching', options: booleanOptions, help: 'Ajoute les panneaux de filtrage par colonne.' },
    { key: 'searchPanesAutomatic', type: 'select', section: 'options', group: 'search', dependsOn: 'searchPanes', options: booleanOptions, help: 'En automatique, DataTables choisit les colonnes selon le seuil de tolérance.' },
    { key: 'searchPanesThreshold', type: 'range', section: 'options', group: 'search', dependsOn: [{ field: 'searchPanes', value: 'true' }, { field: 'searchPanesAutomatic', value: 'true' }], min: 0, max: 1, step: 0.05, placeholder: '0.6', help: 'Seuil entre 0 et 1. Plus il est haut, plus les colonnes avec beaucoup de valeurs uniques restent filtrables.' },
    { key: 'searchPanesColumns', type: 'datatable-columns-select', section: 'options', group: 'search', dependsOn: [{ field: 'searchPanes', value: 'true' }, { field: 'searchPanesAutomatic', value: 'false' }], help: 'Colonnes à afficher dans les panneaux de filtre.' },
    { key: 'paging', type: 'select', section: 'options', group: 'pagination', exclusiveWith: 'scroller', options: booleanOptions, help: 'Active la pagination.' },
    { key: 'pageLength', type: 'number', section: 'options', group: 'pagination', dependsOn: 'paging', placeholder: '25', help: 'Nombre de lignes affichées par page.' },
    { key: 'info', type: 'select', section: 'options', group: 'pagination', dependsOn: 'paging', options: booleanOptions, help: 'Affiche les informations de pagination.' },
    { key: 'ordering', type: 'select', section: 'options', group: 'sorting', options: booleanOptions, help: 'Active le tri sur les colonnes.' },
    { key: 'orderColumn', type: 'datatable-column-select', section: 'options', group: 'sorting', dependsOn: 'ordering', placeholder: '1', help: 'Colonne utilisée pour le tri initial.' },
    { key: 'orderDirection', type: 'select', section: 'options', group: 'sorting', dependsOn: 'ordering', options: orderDirectionOptions, help: 'Sens du tri initial.' },
    { key: 'responsive', type: 'select', section: 'options', group: 'layout', options: booleanOptions, help: 'Adapte les colonnes aux petits écrans.' },
    { key: 'buttons', type: 'select', section: 'options', group: 'exports', options: dataTableButtonOptions, help: 'Ajoute les boutons de copie, export, impression et visibilité des colonnes.' },
    { key: 'colReorder', type: 'select', section: 'options', group: 'columns', options: booleanOptions, help: 'Permet de réordonner les colonnes.' },
    { key: 'fixedHeader', type: 'select', section: 'options', group: 'columns', options: booleanOptions, help: 'Garde l’en-tête visible pendant le scroll.' },
    { key: 'scrollX', type: 'select', section: 'options', group: 'columns', options: booleanOptions, help: 'Active le défilement horizontal.' },
    { key: 'fixedColumnsLeft', type: 'number', section: 'options', group: 'columns', dependsOn: 'scrollX', placeholder: '0', help: 'Nombre de colonnes figées à gauche.' },
    { key: 'fixedColumnsRight', type: 'number', section: 'options', group: 'columns', dependsOn: 'scrollX', placeholder: '0', help: 'Nombre de colonnes figées à droite.' },
    { key: 'rowGroupColumn', type: 'datatable-column-select', section: 'options', group: 'rows', placeholder: 'Désactivé', help: 'Colonne à regrouper. Sélectionnez une colonne pour activer le regroupement de lignes.' },
    { key: 'selectRows', type: 'select', section: 'options', group: 'rows', options: booleanOptions, help: 'Permet la sélection de lignes.' },
    { key: 'scroller', type: 'select', section: 'options', group: 'rows', exclusiveWith: 'paging', options: booleanOptions, help: 'Optimise l’affichage des grandes listes.' },
    { key: 'stateSave', type: 'select', section: 'options', group: 'state', options: booleanOptions, help: 'Mémorise recherche, tri, page et colonnes dans le navigateur.' },
    ...templateVariableFields
  ],
  fieldGroups: {
    search: 'Recherche',
    pagination: 'Pagination',
    sorting: 'Tri initial',
    layout: 'Adaptation',
    exports: 'Exports',
    columns: 'Colonnes et défilement',
    rows: 'Lignes',
    state: 'Persistance'
  },
  defaults: {
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
    searchPanesAutomatic: 'true',
    searchPanesThreshold: 0.6,
    searchPanesColumns: '[]',
    colReorder: 'true',
    fixedHeader: 'true',
    stateSave: 'true',
    selectRows: 'false',
    scrollX: 'false',
    scroller: 'false',
    fixedColumnsLeft: 0,
    fixedColumnsRight: 0,
    rowGroupColumn: '',
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
      searchPanesAutomatic: booleanOption(config.searchPanesAutomatic, this.defaults.searchPanesAutomatic),
      searchPanesThreshold: numberOption(config.searchPanesThreshold, this.defaults.searchPanesThreshold, 0, 1),
      searchPanesColumns: JSON.stringify(normalizeStringArray(config.searchPanesColumns)),
      colReorder: booleanOption(config.colReorder, this.defaults.colReorder),
      fixedHeader: booleanOption(config.fixedHeader, this.defaults.fixedHeader),
      stateSave: booleanOption(config.stateSave, this.defaults.stateSave),
      selectRows: booleanOption(config.selectRows, this.defaults.selectRows),
      scrollX: booleanOption(config.scrollX, this.defaults.scrollX),
      scroller: booleanOption(config.scroller, this.defaults.scroller),
      fixedColumnsLeft: integerOption(config.fixedColumnsLeft, this.defaults.fixedColumnsLeft, 0),
      fixedColumnsRight: integerOption(config.fixedColumnsRight, this.defaults.fixedColumnsRight, 0),
      rowGroupColumn: String(config.rowGroupColumn ?? this.defaults.rowGroupColumn).trim(),
      orderColumn: String(config.orderColumn ?? this.defaults.orderColumn).trim() || String(this.defaults.orderColumn),
      orderDirection,
      globalVariables: normalizeTemplateVariables(config.globalVariables),
      loopVariables: normalizeTemplateVariables(config.loopVariables)
    };
  },
  compile(config, context = {}) {
    const normalized = this.normalizeConfig(config);
    if (!normalized.items) {
      return {
        template: '<div class="text-secondary small">Veuillez configurer la source de données (items)</div>',
        javascript: ''
      };
    }

    let parsedColumns = [];
    if (normalized.columns) {
      try {
        parsedColumns = JSON.parse(normalized.columns);
      } catch {
        parsedColumns = [];
      }
    }
    if (!Array.isArray(parsedColumns)) {
      parsedColumns = [];
    }

    const buttonSets = {
      none: [],
      basic: ['copy', 'csv', 'excel', 'print'],
      all: ['copy', 'csv', 'excel', 'pdf', 'print', 'colvis']
    };
    const searching = toBoolean(normalized.searching);
    const ordering = toBoolean(normalized.ordering);
    const scroller = toBoolean(normalized.scroller) && !toBoolean(normalized.paging);
    const paging = toBoolean(normalized.paging) || scroller;
    const showPagingControls = paging && !scroller;
    const scrollX = toBoolean(normalized.scrollX);
    const responsive = toBoolean(normalized.responsive) && !scrollX;
    const searchBuilder = searching && toBoolean(normalized.searchBuilder);
    const searchPanes = searching && toBoolean(normalized.searchPanes);
    const searchPanesAutomatic = toBoolean(normalized.searchPanesAutomatic);
    const searchPanesColumns = normalizeStringArray(normalized.searchPanesColumns);
    const topStartLayout = [
      searchBuilder ? 'searchBuilder' : '',
      searchPanes ? 'searchPanes' : '',
      normalized.buttons !== 'none' ? 'buttons' : ''
    ].filter(Boolean);
    const topEndLayout = [
      showPagingControls ? 'pageLength' : '',
      searching ? 'search' : ''
    ].filter(Boolean);
    const lengthMenuValues = Array.from(new Set([10, 25, 50, 100, normalized.pageLength, -1]))
      .filter((value) => value === -1 || value > 0)
      .sort((a, b) => (a === -1 ? 1 : b === -1 ? -1 : a - b));
    const lengthMenuLabels = lengthMenuValues.map((value) => value === -1 ? 'All' : value);
    const numericOrderColumn = Number.parseInt(normalized.orderColumn, 10);
    const namedOrderColumnIndex = parsedColumns.findIndex((column) => (
      String(column?.key ?? '') === normalized.orderColumn ||
      String(column?.label ?? '') === normalized.orderColumn
    ));
    const orderColumnIndex = namedOrderColumnIndex >= 0
      ? namedOrderColumnIndex
      : Math.max((Number.isFinite(numericOrderColumn) ? numericOrderColumn : 1) - 1, 0);

    const numericRowGroupColumn = Number.parseInt(normalized.rowGroupColumn, 10);
    const namedRowGroupColumnIndex = parsedColumns.findIndex((column) => (
      String(column?.key ?? '') === normalized.rowGroupColumn ||
      String(column?.label ?? '') === normalized.rowGroupColumn
    ));
    const rowGroupColumnIndex = namedRowGroupColumnIndex >= 0
      ? namedRowGroupColumnIndex
      : (Number.isFinite(numericRowGroupColumn) && numericRowGroupColumn > 0 ? numericRowGroupColumn - 1 : null);

    const fixedColumns = scrollX && (normalized.fixedColumnsLeft || normalized.fixedColumnsRight)
      ? {
           leftColumns: normalized.fixedColumnsLeft,
           rightColumns: normalized.fixedColumnsRight
        }
      : false;
    const dataTableOptions = {
      layout: {
        topStart: topStartLayout.length > 0 ? topStartLayout : null,
        topEnd: topEndLayout.length > 0 ? topEndLayout : null,
        bottomStart: paging && toBoolean(normalized.info) ? 'info' : null,
        bottomEnd: showPagingControls ? 'paging' : null
      },
      pageLength: normalized.pageLength,
      lengthMenu: [lengthMenuValues, lengthMenuLabels],
      lengthChange: showPagingControls,
      searching,
      ordering,
      paging,
      info: paging && toBoolean(normalized.info),
      stateSave: toBoolean(normalized.stateSave),
      order: ordering ? [[orderColumnIndex, normalized.orderDirection]] : [],
      buttons: buttonSets[normalized.buttons]
    };
    if (responsive) {
      dataTableOptions.responsive = true;
    }
    if (toBoolean(normalized.colReorder)) {
      dataTableOptions.colReorder = true;
    }
    if (toBoolean(normalized.fixedHeader)) {
      dataTableOptions.fixedHeader = true;
    }
    if (toBoolean(normalized.selectRows)) {
      dataTableOptions.select = {
        style: 'multi',
        items: 'row'
      };
    }
    if (scrollX) {
      dataTableOptions.scrollX = true;
    }
    if (scroller) {
      dataTableOptions.scroller = true;
      dataTableOptions.deferRender = true;
      dataTableOptions.scrollY = '50vh';
      dataTableOptions.scrollCollapse = true;
    }
    if (searchBuilder) {
      dataTableOptions.searchBuilder = true;
    }
    if (searchPanes) {
      dataTableOptions.searchPanes = searchPanesAutomatic
        ? { threshold: normalized.searchPanesThreshold }
        : { columns: [], threshold: 1 };
    }
    if (fixedColumns) {
      dataTableOptions.fixedColumns = fixedColumns;
    }
    if (rowGroupColumnIndex !== null) {
      dataTableOptions.rowGroup = { dataSrc: rowGroupColumnIndex };
    } else {
      dataTableOptions.rowGroup = false;
    }
    const itemTokenSeed = context.item?.id || context.item?.name
      ? { id: context.item?.id ?? '', name: context.item?.name ?? '' }
      : normalized;
    const tableToken = `dc-dt-${hashString(JSON.stringify(itemTokenSeed))}`;
    const optionsJson = JSON.stringify(dataTableOptions);
    const orderColumnJson = JSON.stringify(normalized.orderColumn);
    const orderDirectionJson = JSON.stringify(normalized.orderDirection);
    const rowGroupColumnJson = JSON.stringify(normalized.rowGroupColumn);
    const searchPanesColumnsJson = JSON.stringify(searchPanes && !searchPanesAutomatic ? searchPanesColumns : []);
    const globalSetup = compileTemplateVariables(normalized.globalVariables);
    const loopSetup = compileTemplateVariables(normalized.loopVariables, normalized.globalVariables, 'loop');

    let templateHtml = '';

    const tableClasses = [
      'table',
      'table-sm',
      'table-striped',
      'table-hover',
      'align-middle',
      'w-100',
      scrollX ? 'text-nowrap' : ''
    ].filter(Boolean).join(' ');

    if (parsedColumns.length > 0) {
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
      templateHtml = `{% if ${normalized.items} %}<table class="${tableClasses}" data-dc-datatable="${tableToken}"><thead><tr>${theadRows}</tr></thead><tbody>${collectionTemplate(normalized.items, rowHtml, rowHtml, `<tr><td colspan="${parsedColumns.length}">Aucune donnée</td></tr>`, normalized.loopVariables, normalized.globalVariables)}</tbody></table>{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    } else {
      const arrayTable = `<table class="${tableClasses}" data-dc-datatable="${tableToken}"><thead><tr>{% for key, value in ${normalized.items}[0] %}<th>{{ key }}</th>{% endfor %}</tr></thead><tbody>{% for item in ${normalized.items} %}{% set key = loop.index0 %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}<tr>{% for key, value in item %}<td>{{ value }}</td>{% endfor %}</tr>{% endfor %}</tbody></table>`;
      const objectTable = `<table class="${tableClasses}" data-dc-datatable="${tableToken}">{% for key, item in ${normalized.items} %}{% set value = item %}{% set _dcRowValue = item %}${loopSetup}{% if loop.first %}<thead><tr><th>Clé</th>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<th>{{ fieldKey }}</th>{% endfor %}{% else %}<th>Valeur</th>{% endif %}</tr></thead><tbody>{% endif %}<tr><td>{{ key }}</td>{% if item is mapping %}{% for fieldKey, fieldValue in item %}<td>{{ fieldValue }}</td>{% endfor %}{% else %}<td>{{ value }}</td>{% endif %}</tr>{% else %}<tbody><tr><td>Aucune donnée</td></tr>{% endfor %}</tbody></table>`;
      templateHtml = `{% if ${normalized.items} and ${normalized.items}.length %}${arrayTable}{% elif ${normalized.items} %}${objectTable}{% else %}<div class="text-secondary">Aucune donnée</div>{% endif %}`;
    }

    return {
      template: `${globalSetup}${templateHtml}`,
      javascript: `const dataTableOptions = ${optionsJson};
const dataTableOrderColumn = ${orderColumnJson};
const dataTableOrderDirection = ${orderDirectionJson};
const dataTableRowGroupColumn = ${rowGroupColumnJson};
const dataTableSearchPanesColumns = ${searchPanesColumnsJson};
if (dataTableOptions.buttons?.includes('excel') && DataTable.Buttons?.jszip) {
  DataTable.Buttons.jszip(jszip);
}
if (dataTableOptions.buttons?.includes('pdf') && DataTable.Buttons?.pdfMake) {
  pdfmake.vfs = pdfFonts?.pdfMake?.vfs || pdfFonts?.vfs || pdfmake.vfs;
  DataTable.Buttons.pdfMake(pdfmake);
}
document.querySelectorAll('table[data-dc-datatable="${tableToken}"]').forEach((table) => {
  if (jQuery.fn?.dataTable?.isDataTable(table)) return;
  const tableOptions = { ...dataTableOptions };
  if (!dataTableOptions.scrollX) {
    delete tableOptions.scrollX;
    table.style.width = '100%';
  }
  if (tableOptions.stateSave) {
    tableOptions.stateLoadParams = (_settings, data) => {
      delete data.length;
      delete data.start;
    };
  }
  if (tableOptions.searchPanes && dataTableSearchPanesColumns.length > 0) {
    const selectedColumns = new Set(dataTableSearchPanesColumns.map((column) => String(column).trim().toLowerCase()));
    const selectedIndexes = Array.from(table.querySelectorAll('thead th'))
      .map((header, index) => selectedColumns.has(header.textContent.trim().toLowerCase()) ? index : null)
      .filter((index) => index !== null);
    tableOptions.searchPanes = {
      ...tableOptions.searchPanes,
      columns: selectedIndexes
    };
    tableOptions.columnDefs = [
      ...(tableOptions.columnDefs || []),
      ...selectedIndexes.map((index) => ({
        targets: index,
        searchPanes: { show: true, threshold: 1 }
      }))
    ];
  }
  if (tableOptions.ordering && dataTableOrderColumn && !/^[1-9]\\d*$/.test(String(dataTableOrderColumn))) {
    const normalizedOrderColumn = String(dataTableOrderColumn).trim().toLowerCase();
    const columnAliases = {
      key: ['key', 'clé'],
      value: ['value', 'valeur']
    };
    const expectedHeaders = columnAliases[normalizedOrderColumn] || [normalizedOrderColumn];
    const headerIndex = Array.from(table.querySelectorAll('thead th')).findIndex((header) => (
      expectedHeaders.includes(header.textContent.trim().toLowerCase())
    ));
    if (headerIndex >= 0) {
      tableOptions.order = [[headerIndex, dataTableOrderDirection]];
    }
  }
  if (dataTableRowGroupColumn && !/^[1-9]\\d*$/.test(String(dataTableRowGroupColumn))) {
    const normalizedGroupColumn = String(dataTableRowGroupColumn).trim().toLowerCase();
    const columnAliases = {
      key: ['key', 'clé'],
      value: ['value', 'valeur']
    };
    const expectedHeaders = columnAliases[normalizedGroupColumn] || [normalizedGroupColumn];
    const headerIndex = Array.from(table.querySelectorAll('thead th')).findIndex((header) => (
      expectedHeaders.includes(header.textContent.trim().toLowerCase())
    ));
    if (headerIndex >= 0) {
      tableOptions.rowGroup = { dataSrc: headerIndex };
    } else {
      tableOptions.rowGroup = false;
    }
  } else if (!dataTableRowGroupColumn) {
    tableOptions.rowGroup = false;
  }
  new DataTable(table, tableOptions);
  if (!dataTableOptions.scrollX) {
    const scrollBody = table.closest('.dt-container')?.querySelector('.dt-scroll-body');
    if (scrollBody) {
      scrollBody.style.overflowX = 'hidden';
    }
  }
});`
    };
  }
});
