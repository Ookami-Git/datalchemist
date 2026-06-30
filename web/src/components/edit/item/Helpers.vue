<script setup>
import TemplateVariablesReference from '../common/TemplateVariablesReference.vue';

const props = defineProps({
  sections: { type: Array, default: null },
  context: {
    type: String,
    default: 'item',
    validator: (value) => ['item', 'source'].includes(value)
  }
});

function showSection(section) {
  return !props.sections?.length || props.sections.includes(section);
}

// Nunjucks date examples
const codeDateSyntax = `{{ var_date | date('outputformat','inputformat') }}`
const codeDateExample1 = `{{ "2022-01-10" | date("DD MMM YYYY") }}\n=> 10 jan 2022`
const today = new Date().toISOString().slice(0, 10)
const codeDateExample2 = `{{ now | date('YYYY-MM-DD') }}\n=> ${today} ("now" is undefined var, result is today)`
const codeDateExample3 = `{{ "01/12/2022" | date('YYYY-MM-DD','DD/MM/YYYY') }}\n=> 2022-12-01`

// Nunjucks find
const codeFindSyntax = `{{ var_array | find('key.path', 'value') }}`

// Nunjucks fromjson
const codeFromJsonSyntax = `{{ var_jsonstring | fromjson }}`

// Nunjucks setAttribute
const codeSetAttrSyntax = `{{ var_object | setAttribute('key.path', 'value') }}`
const codeSetAttrExample1 = `{{ {"key1": "value1"} | setAttribute('key2', 'new key/value') }}\n=> {"key1": "value1", "key2": "new value"}`
const codeSetAttrExample2 = `{{ {"key.with.dot": "value1", "key2": "value2"} | setAttribute('key\\\\.with\\\\.dot', 'new value') }}\n=> {"key.with.dot": "new value", "key2": "value2"}`
const codeSetAttrExample3 = `{{ { "key1": {"level1":"value1"} } | setAttribute('key1.level1', {"level2":"new value"}) }}\n=> {"key1":{"level1":{"level2":"new value"}}}`

// Nunjucks split
const codeSplitSyntax = `{{ var_string | split('separator') }}`
const codeSplitExample1 = `{{ "a,b,c" | split(',') }}\n=> ["a", "b", "c"]`
const codeSplitExample2 = `{{ "one two three" | split(' ') }}\n=> ["one", "two", "three"]`

// Icons
const codeBootstrapIcon = `<i class="bi bi-[icon-name]"></i>`
const codeFontawesomeIcon = `<i class="fa fa-[icon-name]"></i>`

// Datatable
const codeDatatableHtml = `<table id="uniquetableid" class="table datatable">...</table>`
const codeDatatableJs = `new DataTable('#uniquetableid', { dom: 'Bfrtip', buttons: ['copy' , 'csv', 'excel'] });`

// Mermaid
const codeMermaid = `<pre class="mermaid"></pre>`
</script>

<style>
@import url("highlight.js/styles/atom-one-dark.min.css");
</style>

<template>
  <div class="card edit-item-helper-card">
    <div class="card-header d-flex flex-column align-items-start gap-1">
      <span class="fw-semibold">
        <i class="bi bi-lightbulb-fill me-1"></i>{{ $t('edititem.global.documentation') }}
      </span>
      <span class="small text-secondary">{{ $t('edititem.global.docsbytype') }}</span>
    </div>
    <div class="card-body p-0">
      <div class="accordion accordion-flush" id="item-helpers-accordion">
        <div v-if="showSection('variables')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-variables-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-variables" aria-expanded="false" aria-controls="item-helper-variables">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">{{ $t('templateVariables.title') }}</span>
                <span class="badge text-bg-secondary">Template</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-variables" class="accordion-collapse collapse"
            aria-labelledby="item-helper-variables-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <TemplateVariablesReference :context="props.context" />
            </div>
          </div>
        </div>

        <div v-if="showSection('bootstrap')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-bootstrap-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-bootstrap" aria-expanded="false" aria-controls="item-helper-bootstrap">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">Bootstrap</span>
                <span class="badge text-bg-secondary">HTML/CSS</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-bootstrap" class="accordion-collapse collapse"
            aria-labelledby="item-helper-bootstrap-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              {{ $t('edititem.bootstrap.description') }}
              <div class="mt-2">
                <a href="https://getbootstrap.com/docs/5.3/getting-started/introduction/"
                  target="_blank">https://getbootstrap.com/docs/5.3/getting-started/introduction/</a>
              </div>
            </div>
          </div>
        </div>

        <div v-if="showSection('icons')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-icons-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-icons" aria-expanded="false" aria-controls="item-helper-icons">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">{{ $t('edititem.icons.header') }}</span>
                <span class="badge text-bg-secondary">HTML/CSS</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-icons" class="accordion-collapse collapse" aria-labelledby="item-helper-icons-heading"
            data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <p class="mb-2">{{ $t('edititem.icons.description') }}</p>
              <p class="fw-semibold mb-2">{{ $t('edititem.global.syntax') }}</p>
              <div class="mb-2">Bootstrap</div>
              <highlightjs language="html" :code="codeBootstrapIcon" />
              <div class="mt-2 mb-2">Fontawesome</div>
              <highlightjs language="html" :code="codeFontawesomeIcon" />
              <ul class="mt-3 mb-0">
                <li><a href="https://icons.getbootstrap.com/" target="_blank">Bootstrap Icons</a></li>
                <li><a href="https://fontawesome.com/search?o=r&m=free" target="_blank">Font Awesome</a></li>
              </ul>
            </div>
          </div>
        </div>

        <div v-if="showSection('datatables')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-datatables-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-datatables" aria-expanded="false" aria-controls="item-helper-datatables">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">DataTables</span>
                <span class="badge text-bg-secondary">HTML/JS</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-datatables" class="accordion-collapse collapse"
            aria-labelledby="item-helper-datatables-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <p class="mb-2">{{ $t('edititem.datatable.description') }}</p>
              <div class="fw-semibold mb-2">HTML</div>
              <highlightjs language="html" :code="codeDatatableHtml" />
              <div class="fw-semibold mt-2 mb-2">JavaScript</div>
              <highlightjs language="javascript" :code="codeDatatableJs" />
              <div class="mt-2">
                <a href="https://datatables.net/examples/index"
                  target="_blank">https://datatables.net/examples/index</a>
              </div>
            </div>
          </div>
        </div>

        <div v-if="showSection('mermaid')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-mermaid-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-mermaid" aria-expanded="false" aria-controls="item-helper-mermaid">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">Mermaid</span>
                <span class="badge text-bg-secondary">Graphs</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-mermaid" class="accordion-collapse collapse"
            aria-labelledby="item-helper-mermaid-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <p class="mb-2">{{ $t('edititem.mermaid.description') }}</p>
              <highlightjs language="html" :code="codeMermaid" />
              <div class="mt-2">
                <a href="https://mermaid.js.org/intro/" target="_blank">https://mermaid.js.org/intro/</a>
              </div>
            </div>
          </div>
        </div>

        <div v-if="showSection('nunjucks')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-nunjucks-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-nunjucks" aria-expanded="false" aria-controls="item-helper-nunjucks">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">Nunjucks</span>
                <span class="badge text-bg-secondary">Template</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-nunjucks" class="accordion-collapse collapse"
            aria-labelledby="item-helper-nunjucks-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <p class="mb-2">{{ $t('edititem.nunjucks.description') }}</p>
              <div class="mb-3">
                <a href="https://mozilla.github.io/nunjucks/templating.html" target="_blank">
                  https://mozilla.github.io/nunjucks/templating.html
                </a>
              </div>
              <p class="fw-semibold mb-2">{{ $t('edititem.nunjucks.customfilter') }}</p>

              <div class="mb-3">
                <div><code>date</code> - {{ $t('edititem.nunjucks.date.description') }}</div>
                <div class="small text-secondary mb-1">{{ $t('edititem.global.syntax') }}: <a
                    href="https://momentjs.com/docs/#/displaying/format/" target="_blank">MomentJS format</a></div>
                <highlightjs language="nunjucks" :code="codeDateSyntax" />
                <div class="mt-2 small text-secondary">{{ $t('edititem.global.examples') }}</div>
                <highlightjs language="nunjucks" :code="codeDateExample1" />
                <highlightjs language="nunjucks" :code="codeDateExample2" />
                <highlightjs language="nunjucks" :code="codeDateExample3" />
              </div>

              <div class="mb-3">
                <div><code>find</code> - {{ $t('edititem.nunjucks.find.description') }}</div>
                <highlightjs language="nunjucks" :code="codeFindSyntax" />
              </div>

              <div class="mb-3">
                <div><code>fromjson</code> - {{ $t('edititem.nunjucks.fromjson.description') }}</div>
                <highlightjs language="nunjucks" :code="codeFromJsonSyntax" />
              </div>

              <div class="mb-3">
                <div><code>setAttribute</code> - {{ $t('edititem.nunjucks.setattribute.description') }}</div>
                <highlightjs language="nunjucks" :code="codeSetAttrSyntax" />
                <div class="mt-2 small text-secondary">{{ $t('edititem.global.examples') }}</div>
                <highlightjs language="nunjucks" :code="codeSetAttrExample1" />
                <highlightjs language="nunjucks" :code="codeSetAttrExample2" />
                <highlightjs language="nunjucks" :code="codeSetAttrExample3" />
              </div>

              <div>
                <div><code>split</code> - {{ $t('edititem.nunjucks.split.description') }}</div>
                <highlightjs language="nunjucks" :code="codeSplitSyntax" />
                <div class="mt-2 small text-secondary">{{ $t('edititem.global.examples') }}</div>
                <highlightjs language="nunjucks" :code="codeSplitExample1" />
                <highlightjs language="nunjucks" :code="codeSplitExample2" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="showSection('javascript')" class="accordion-item">
          <h2 class="accordion-header" id="item-helper-javascript-heading">
            <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
              data-bs-target="#item-helper-javascript" aria-expanded="false" aria-controls="item-helper-javascript">
              <span class="d-flex align-items-center gap-2">
                <span class="fw-semibold">JavaScript</span>
                <span class="badge text-bg-secondary">JS</span>
              </span>
            </button>
          </h2>
          <div id="item-helper-javascript" class="accordion-collapse collapse"
            aria-labelledby="item-helper-javascript-heading" data-bs-parent="#item-helpers-accordion">
            <div class="accordion-body">
              <p class="mb-2">{{ $t('edititem.javascript.description') }}</p>
              <a href="https://www.w3schools.com/js/default.asp"
                target="_blank">https://www.w3schools.com/js/default.asp</a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
