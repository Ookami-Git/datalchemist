<script setup>
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
  <div class="card">
    <div class="card-body">
      <!-- BOOTSTRAP -->
      <a href="https://getbootstrap.com/docs/5.3/getting-started/introduction/" target="_blank">Bootstrap</a> (html/css) <a data-bs-toggle="collapse" href="#collapseBootstrap"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseBootstrap">
        <div class="card card-body">
          {{ $t('edititem.bootstrap.description') }}
        </div>
      </div>
      <br>
      <!-- ICONS -->
      <span class="text-capitalize">{{ $t('edititem.icons.header') }}</span> (html/css) <a data-bs-toggle="collapse" href="#collapseIcons"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseIcons">
        <div class="card card-body">
          {{ $t('edititem.icons.description') }}  <br>
          <span class="text-capitalize">{{ $t('edititem.global.syntax') }} :</span> 
            <ul>
              <li>
                Bootstrap : 
                <highlightjs language="html" :code="codeBootstrapIcon" />
              </li>
              <li>
                Fontawesome : 
                <highlightjs language="html" :code="codeFontawesomeIcon" />
              </li>
            </ul>
          {{ $t('edititem.icons.header') }} :
            <ul>
              <li><a href="https://icons.getbootstrap.com/" target="_blank">Bootstrap</a></li>
              <li><a href="https://fontawesome.com/search?o=r&m=free" target="_blank">Fontawesome</a></li>
            </ul>
        </div>
      </div>
      <br>
      <!-- Datatable -->
      <a href="https://datatables.net/examples/index" target="_blank">DataTables</a> (html/js) <a data-bs-toggle="collapse" href="#collapseDatatable"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseDatatable">
        <div class="card card-body">
          {{ $t('edititem.datatable.description') }}<br>
          {{ $t('edititem.global.syntax') }} : <br><br>
          <ul>
            <li>
              HTML :<br>
              <highlightjs language="html" :code="codeDatatableHtml" />
            </li>
            <li>
              Javascript :<br>
              <highlightjs language="javascript" :code="codeDatatableJs" />
            </li>
          </ul>
        </div>
      </div>
      <br>
      <!-- MERMAID -->
      <a href="https://mermaid.js.org/intro/" target="_blank">Mermaid</a> (Graphs) <a data-bs-toggle="collapse" href="#collapseMermaid"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseMermaid">
        <div class="card card-body">
          {{ $t('edititem.mermaid.description') }}<br>
          {{ $t('edititem.global.syntax') }} : 
          <highlightjs language="html" :code="codeMermaid" />
        </div>
      </div>
      <br>
      <!-- NUNJUCKS -->
      <a href="https://mozilla.github.io/nunjucks/fr/templating.html" target="_blank">Nunjucks</a> (Template) <a data-bs-toggle="collapse" href="#collapseNunjucks"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseNunjucks">
        <div class="card card-body">
          {{ $t('edititem.nunjucks.description') }}<br>
          {{ $t('edititem.nunjucks.customfilter') }} : <br>
          <ul>
            <li>
              <code>date</code> : {{ $t('edititem.nunjucks.date.description') }}<br>
              {{ $t('edititem.global.syntax') }} format : <a href="https://momentjs.com/docs/#/displaying/format/" target="_blank">Momentjs format</a><br>
              {{ $t('edititem.global.syntax') }} : 
              <highlightjs language="nunjucks" :code="codeDateSyntax" />
              <br>
              {{ $t('edititem.global.examples') }} : 
              <ul>
                <li>
                  <highlightjs language="nunjucks" :code="codeDateExample1" />
                </li>
                <li>
                  <highlightjs language="nunjucks" :code="codeDateExample2" />
                </li>
                <li>
                  <highlightjs language="nunjucks" :code="codeDateExample3" />
                </li>
              </ul>
            </li>
            <li>
              <code>find</code> : {{ $t('edititem.nunjucks.find.description') }}<br>
              {{ $t('edititem.global.syntax') }} : 
              <highlightjs language="nunjucks" :code="codeFindSyntax" />
            </li>
            <li>
              <code>fromjson</code> : {{ $t('edititem.nunjucks.fromjson.description') }}<br>
              {{ $t('edititem.global.syntax') }} : 
              <highlightjs language="nunjucks" :code="codeFromJsonSyntax" />
            </li>
            <li>
              <code>setAttribute</code> : {{ $t('edititem.nunjucks.setattribute.description') }}<br>
              {{ $t('edititem.global.syntax') }} : 
              <highlightjs language="nunjucks" :code="codeSetAttrSyntax" />
              <br>
              {{ $t('edititem.global.examples') }} : 
              <ul>
                <li>
                  <highlightjs language="nunjucks" :code="codeSetAttrExample1" />
                </li>
                <li>
                  <highlightjs language="nunjucks" :code="codeSetAttrExample2" />
                </li>
                <li>
                  <highlightjs language="nunjucks" :code="codeSetAttrExample3" />
                </li>
              </ul>
            </li>
            <li>
              <code>split</code> : {{ $t('edititem.nunjucks.split.description') }}<br>
              {{ $t('edititem.global.syntax') }} : 
              <highlightjs language="nunjucks" :code="codeSplitSyntax" />
              <br>
              {{ $t('edititem.global.examples') }} :
              <ul>
                <li>
                  <highlightjs language="nunjucks" :code="codeSplitExample1" />
                </li>
                <li>
                  <highlightjs language="nunjucks" :code="codeSplitExample2" />
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
      <br>
      <!-- Javascript -->
      <a href="https://www.w3schools.com/js/default.asp" target="_blank">Javascript</a> (js) <a data-bs-toggle="collapse" href="#collapseJavascript"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseJavascript">
        <div class="card card-body">
          {{ $t('edititem.javascript.description') }}
        </div>
      </div>
      <br>
    </div>
  </div>
</template>