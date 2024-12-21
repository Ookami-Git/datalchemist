<script setup>
  const openBrace = '{{'
  const closeBrace = '}}'
</script>

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
              <li>Bootstrap : <code>&lt;i class="bi bi-[icon-name]"&gt;&lt;/i&gt;</code></li>
              <li>Fontawesome : <code>&lt;i class="fa fa-[icon-name]"&gt;&lt;/i&gt;</code></li>
            </ul>
          {{ $t('edititem.icons.header') }} :
            <ul>
              <li><a href="https://icons.getbootstrap.com/" target="_blank">Bootstrap</a></li>
              <li><a href="https://fontawesome.com/search?o=r&m=free" target="_blank">Fontawesome</a></li>
            </ul>
        </div>
      </div>
      <br>
      <!-- Table -->
      <span>{{ $t('edititem.table.header') }}</span> (html/css) <a data-bs-toggle="collapse" href="#collapseTable"><i class="bi bi-caret-down-square-fill"></i></a>
      <div class="collapse" id="collapseTable">
        <div class="card card-body">
          <ul>
            <li>
              {{ $t('edititem.table.sortable.header') }} :
              {{ $t('edititem.table.sortable.description') }}<br><a href="https://github.com/tofsjonas/sortable/blob/main/README.md" target="_blank">{{ $t('edititem.table.sortable.doc') }}</a><br>
              {{ $t('edititem.global.syntax') }} : <code>&lt;table class="table sortable"&gt;&lt;/table&gt;</code><br>
            </li>
            <li>
              {{ $t('edititem.table.filterable.header') }} :
              {{ $t('edititem.table.filterable.description') }}<br>
              {{ $t('edititem.global.syntax') }} : <code>&lt;table class="table filterable"&gt;&lt;/table&gt;</code><br>
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
          {{ $t('edititem.global.syntax') }} : <code>&lt;pre class="mermaid"&gt;&lt;/pre&gt;</code>
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
              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_date | date("outputformat","inputformat") {{ closeBrace }}</code><br>
              {{ $t('edititem.global.examples') }} : 
              <ul>
                <li><code>{{ openBrace }} "2022-01-10" | date("DD MMM YYYY") {{ closeBrace }}</code> : 10 jan 2022</li>
                <li><code>{{ openBrace }} now | date("YYYY-MM-DD") {{ closeBrace }}</code> : 2024-01-01 ("now" is undefined var, result is today)</li>
                <li><code>{{ openBrace }} "01/12/2022" | date("YYYY-MM-DD","DD/MM/YYYY") {{ closeBrace }}</code> : 2022-12-01</li>
              </ul>
            </li>
            <li>
              <code>find</code> : {{ $t('edititem.nunjucks.find.description') }}<br>
              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_array | find("key.path", "value") {{ closeBrace }}</code>
            </li>
            <li>
              <code>fromjson</code> : {{ $t('edititem.nunjucks.fromjson.description') }}<br>
              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_jsonstring | fromjson {{ closeBrace }}</code>
            </li>
            <li>
              <code>setAttribute</code> : {{ $t('edititem.nunjucks.setattribute.description') }}<br>
              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_object | setAttribute("key.path", "value") {{ closeBrace }}</code><br>
              {{ $t('edititem.global.examples') }} : 
              <ul>
                <li><code>{{ openBrace }} {"key1": "value1", "key2": "value2"} | setAttribute("key2", "new value") | dump {{ closeBrace }}</code> :<br> {"key1": "value1", "key2": "new value"}</li>
                <li><code>{{ openBrace }} {"key.with.dot": "value1", "key2": "value2"} | setAttribute("key\\.with\\.dot", "new value") | dump {{ closeBrace }}</code> :<br> {"key.with.dot": "new value", "key2": "value2"}</li>
                <li><code>{{ openBrace }} { "key1": {"level1":"value1"} } | setAttribute("key1.level1", {"level2":"new value"}) | dump {{ closeBrace }}</code> :<br> {"key1":{"level1":{"level2":"new value"}}}</li>
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