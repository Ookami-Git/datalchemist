<script setup>
import { ref, onMounted, onUnmounted, inject } from 'vue';
import Papa from 'papaparse';

const tables = ref([]);

const parameters = inject('parameters');

onMounted(() => {
    const updateTables = () => {
        tables.value = [];
        const foundTables = document.querySelectorAll("table.exportable");
        foundTables.forEach((table, index) => {
            const tableId = table.id ? table.id : `table_${index + 1}`;
            tables.value.push({ id: tableId, element: table });
        });
    };

    // Initial update
    updateTables();

    // Observe DOM changes
    const observer = new MutationObserver(updateTables);
    observer.observe(document.body, { childList: true, subtree: true });

    // Cleanup observer on unmount
    onUnmounted(() => {
        observer.disconnect();
    });
});

function download(table, name) {
    const rows = table.rows;
    const headers = [];
    const jsonData = [];
    const ignoreColumns = [];

    // Extract headers
    for (let i = 0; i < rows[0].cells.length; i++) {
        const cell = rows[0].cells[i];
        if (!cell.classList.contains('export-ignore')) {
            headers.push(cell.innerText);
        } else {
            ignoreColumns.push(i);
        }
    }

    // Extract data
    for (let i = 1; i < rows.length; i++) {
        const rowObject = {};
        const cells = rows[i].cells;
        for (let j = 0; j < cells.length; j++) {
            if (!ignoreColumns.includes(j)) {
                const cell = cells[j];
                rowObject[headers[j - ignoreColumns.length]] = cell.getAttribute('data-export-value') || cell.innerText;
            }
        }
        jsonData.push(rowObject);
    }

    var papaparams = {
        quotes: false, //or array of booleans
        quoteChar: '"',
        escapeChar: '"',
        delimiter: parameters.value.export_csv_delimiter,
        header: true,
        newline: "\r\n",
        skipEmptyLines: false, //other option is 'greedy', meaning skip delimiters, quotes, and whitespace.
        columns: null //or array of strings
    };

    var blob = new Blob([Papa.unparse(jsonData, papaparams)], { type: 'text/csv;charset=utf-8;' });

    var link = document.createElement("a");

    var url = URL.createObjectURL(blob);
    link.setAttribute("href", url);
    link.setAttribute("download", name + '.csv');
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}
</script>

<template>
    <div v-if="tables.length > 0" class="dropdown">
        <button class="btn dropdown-toggle" type="button" id="dropdownMenuButton" data-bs-toggle="dropdown" aria-expanded="false">
            <i class="bi bi-cloud-arrow-down-fill"></i>
        </button>
        <ul class="dropdown-menu" aria-labelledby="dropdownMenuButton">
            <li v-for="(table, index) in tables" :key="index">
                <a class="dropdown-item" @click="() => download(table.element, table.id)"><i class="bi bi-file-earmark-spreadsheet"></i> {{ table.id }}</a>
            </li>
        </ul>
    </div>
</template>