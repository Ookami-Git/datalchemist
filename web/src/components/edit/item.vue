<script setup>
import { ref, inject, watch, onMounted, provide } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

import HtmlEditor from "./item/HtmlEditor.vue";
import JSEditor from "./item/JavascriptEditor.vue";
import Helpers from "./item/Helpers.vue";
import sources from "./common/sources.vue";
import Preview from '../view/preview.vue';

const props = defineProps({
  itemid: [String, Number]
});

const typeSource = "item";

const route = useRoute();
const itemid = props.itemid || route.params.itemid;
const save = inject('save');
const apiUrl = inject('apiUrl');
save.value.safe();
const ItemInfo = ref(null);

const parameter = inject('parameters');
// Preview modal
const showPreview = ref(false);
function openPreview() { showPreview.value = true; }
function closePreview() { showPreview.value = false; }
const code = ref("<-- HTML Code -->");
const codeJs = ref("// Javascript Code");

provide('codeHtml', code);
provide('codeJs', codeJs);

const fetchItem = async () => {
  axios.get(`${apiUrl}/item/${itemid}`)
    .then(function (response) {
      code.value = response.data.template;
      codeJs.value = response.data.javascript;
      ItemInfo.value = response.data;
    })
    .catch(function (error) {
      code.value = error;
      codeJs.value = error;
      console.error(`Error fetching data for item ${itemid}`, error);
    });
};

/**
 * Updates an item on the server by sending a POST request to the '/item' endpoint.
 *
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function updateItem() {
  axios.post(`${apiUrl}/item`, {
    id: ItemInfo.value.id,
    name: ItemInfo.value.name,
    template: code.value,
    javascript: codeJs.value
  })
    .then(function () {
      save.value.status.show()
    })
    .catch(function (error) {
      console.log(error);
      save.value.status.error()
    });
}

watch([code, codeJs, ItemInfo], () => {
  if (save.value.show && (code.value !== ItemInfo.value.template || codeJs.value !== ItemInfo.value.javascript)) {
    save.value.status.saveable()
  }
}, { deep: true });

onMounted(async () => {
  await fetchItem(itemid);
  save.value.function = updateItem
  save.value.status.show()
  document.querySelectorAll('[data-bs-toggle="tab"]').forEach((tab) => {
    tab.addEventListener('shown.bs.tab', () => {
      refreshCodeMirror(); // Refresh code mirror on tab change
    });
  });
})

// Fonction for refreshing CodeMirror
const refreshCodeMirror = () => {
  document.querySelectorAll('.CodeMirror').forEach((el) => {
    el.CodeMirror.refresh();
  });
};
</script>

<template>
  <div class="row">
    <div class="col-md-12">
      <div class="card">
        <div class="card-header">
          <div class="row">
            <div class="col-md-1">
              <RouterLink type="button" class="btn btn-secondary btn-sm me-2" :to="{ name: 'edit' }"
                active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
              <button v-if="ItemInfo" class="btn btn-outline-info btn-sm" @click="openPreview" title="Live Preview"><i
                  class="bi bi-eye"></i></button>
            </div>
            <div v-if="ItemInfo" class="col-md-10 text-center">
              <div class="input-group">
                <span class="input-group-text" id="viewname">{{ $t('edititem.header') }}</span>
                <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ ItemInfo.id
                }}</span>
                <input type="text" class="form-control" placeholder="Name" aria-label="View Name"
                  aria-describedby="viewname" v-model="ItemInfo.name">
              </div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <!-- TAB Selector -->
          <div class="row">
            <div class="col-md-8">
              <ul class="nav nav-tabs" id="myTab" role="tablist">
                <li class="nav-item" role="presentation">
                  <button class="nav-link active" id="html-tab" data-bs-toggle="tab" data-bs-target="#html-tab-pane"
                    type="button" role="tab" aria-controls="html-tab-pane" aria-selected="true">HTML</button>
                </li>
                <li class="nav-item" role="presentation">
                  <button class="nav-link" id="javascript-tab" data-bs-toggle="tab"
                    data-bs-target="#javascript-tab-pane" type="button" role="tab" aria-controls="javascript-tab-pane"
                    aria-selected="false">Javascript</button>
                </li>
              </ul>
            </div>
          </div>
          <!-- Editor Page -->
          <div class="row">
            <!-- Editor column -->
            <div class="col-md-8">
              <!-- HTML tab -->
              <div class="tab-content" id="code-tab-content" v-if="parameter.name">
                <div class="tab-pane fade show active row" id="html-tab-pane" role="tabpanel" aria-labelledby="html-tab"
                  tabindex="0">
                  <HtmlEditor />
                </div>
                <!-- JS tab -->
                <div class="tab-pane fade" id="javascript-tab-pane" role="tabpanel" aria-labelledby="javascript-tab"
                  tabindex="0">
                  <JSEditor />
                </div>
              </div>
            </div>
            <!-- Sources & Exlpers column -->
            <div class="col-md-4">
              <!-- Helpers -->
              <Helpers />
              <br>
              <!-- Sources -->
              <sources :typeSource="typeSource" :parentId="itemid" />
            </div>
          </div>
          <br>
        </div>
      </div>
    </div>
  </div>
  <!-- Modal Preview -->
  <div class="modal fade" tabindex="-1" :class="{ show: showPreview }"
    :style="{ display: showPreview ? 'block' : 'none', background: 'rgba(0,0,0,0.3)' }" @click.self="closePreview">
    <div class="modal-dialog modal-xl">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Live Preview</h5>
          <button type="button" class="btn-close" @click="closePreview"></button>
        </div>
        <div class="modal-body">
          <Preview v-if="showPreview" :itemid="itemid" :mode="'edit'" :item="{
            title: `Preview - ${ItemInfo.name}`,
            template: code,
            javascript: codeJs
          }" />
        </div>
      </div>
    </div>
  </div>
</template>
