// --- DataTables & Dépendances ---
import DataTable from 'datatables.net-bs5';
import languageFr from 'datatables.net-plugins/i18n/fr-FR.mjs';
import languageEn from 'datatables.net-plugins/i18n/en-GB.mjs';
import 'datatables.net-buttons-bs5';
import 'datatables.net-buttons/js/buttons.colVis.mjs';
import 'datatables.net-buttons/js/buttons.html5.mjs';
import 'datatables.net-buttons/js/buttons.print.mjs';
import 'datatables.net-colreorder-bs5';
import 'datatables.net-fixedcolumns-bs5';
import 'datatables.net-fixedheader-bs5';
import 'datatables.net-responsive-bs5';
import 'datatables.net-rowgroup-bs5';
import 'datatables.net-scroller-bs5';
import 'datatables.net-searchbuilder-bs5';
import 'datatables.net-searchpanes-bs5';
import 'datatables.net-select-bs5';

export function setDataTablesLanguage(lang) {
  let langModule;
  switch (lang) {
    case 'fr': langModule = languageFr; break;
    case 'en': langModule = languageEn; break;
    default: langModule = languageEn; break;
  }
  Object.assign(DataTable.defaults, {
    language: langModule
  });
}