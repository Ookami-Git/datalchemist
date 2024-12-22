<script setup>
import { inject, watch, onMounted, ref } from 'vue';
import axios from 'axios';

const parameter = ref(null)
const original_parameter = {};
const apiUrl = inject('apiUrl');
const save = inject('save');
save.value.safe()
const views = ref(null);

function SaveParameters () {
    let save_error = false
    for (const [key, value] of Object.entries(parameter.value)) {
        if (value == original_parameter.value[key]) {
            continue;
        }
        axios.put(`${apiUrl}/parameter/${key}`, {
            Name: key,
            Value: `${value}`
        })
        .then(function (response) {
            localStorage.setItem('reloadparameters', true);
            original_parameter[key] = value
        })
        .catch(function (error) {
            console.log(error);
            save_error = true
        });
    }
    if (save_error) {
        save.value.status.error()
    } else {
        save.value.status.show()
    }
}

function fetchParameters () {
    axios.get(`${apiUrl}/parameters/admin`)
    .then(function (response) {
        parameter.value = response.data
        original_parameter.value = JSON.parse(JSON.stringify(response.data))
    })
    .catch(function (error) {
        console.log(error);
    });
}

const fetchViews = async () => {
  axios.get(`${apiUrl}/views`)
  .then(function (response) {
    views.value = response.data;
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des vues`, error);
  });
};

watch (parameter, () => {
    for (const [key, value] of Object.entries(parameter.value)) {
        if (value != original_parameter.value[key]) {
            save.value.status.saveable()
            break;
        }
    }
}, { deep: true });

onMounted(() => {
    fetchParameters()
    fetchViews()
    save.value.function = SaveParameters
    save.value.status.show()
})
</script>

<template>
<div class="card">
    <h6 class="card-header text-center">{{ $t('admin.global.header') }}</h6>
    <div class="card-body" v-if="parameter">
        <div class="row">
            <div class="col-md-3">
                <div class="input-group mb-3">
                    <span class="input-group-text">{{ $t('admin.global.name') }}</span>
                    <input type="text" class="form-control" v-model="parameter.name">
                </div>
            </div>
            <div class="col-md-3">
                <div class="input-group mb-3">
                <label class="input-group-text" for="inputSelectTheme">{{ $t('admin.global.theme') }}</label>
                <select class="form-select" id="inputSelectTheme" v-model="parameter.theme">
                    <option value="light">{{ $t('admin.global.light') }}</option>
                    <option value="dark">{{ $t('admin.global.dark') }}</option>
                </select>
                </div>
            </div>
            <div class="col-md-3">
                <div class="input-group mb-3">
                <label class="input-group-text" for="inputSelectLang">{{ $t('admin.global.lang') }}</label>
                <select class="form-select" id="inputSelectLang" v-model="parameter.lang">
                    <option value="en">English</option>
                    <option value="fr">Français</option>
                </select>
                </div>
            </div>
            <div class="col-md-3">
                <div class="input-group mb-3">
                    <label class="input-group-text" for="inputSelectView">{{ $t('admin.global.defaultview') }}</label>
                    <select class="form-select" id="inputSelectView" v-model="parameter.defaultview">
                        <option v-for="view in views" :key="view.id" :value="view.id">{{ view.name }}</option>
                    </select>
                </div>
            </div>
        </div>
        <div class="card">
            <div class="card-header text-center">{{ $t('admin.global.bg') }}</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <div class="input-group input-group-sm">
                            <span class="input-group-text">{{ $t('admin.global.light') }}</span>
                            <input type="color" id="bg_color_light" v-model="parameter.bg_color_light" class="form-control form-control-color">
                            <input type="color" id="bg_color2_light" v-model="parameter.bg_color2_light" class="form-control form-control-color">
                        </div>
                    </div>
                    <div class="col-md-6">
                        <div class="input-group input-group-sm">
                            <span class="input-group-text">{{ $t('admin.global.dark') }}</span>
                            <input type="color" id="bg_color_dark" v-model="parameter.bg_color_dark" class="form-control form-control-color" >
                            <input type="color" id="bg_color2_dark" v-model="parameter.bg_color2_dark" class="form-control form-control-color">
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <br>
        <div class="card">
            <div class="card-header text-center">{{ $t('admin.global.ldap.header') }}</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-4">
                        <div class="form-check form-check-inline">
                            <input class="form-check-input hidden" type="checkbox" tabindex="0" name="[ldapEnable]:boolean" v-model="parameter.ldap">
                            <label>{{ $t('admin.global.ldap.enable') }}</label>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="form-check form-check-inline">
                            <input class="form-check-input hidden" type="checkbox" tabindex="0" name="[ldapSsl]:boolean" v-model="parameter.ldap_ssl">
                            <label>SSL (LDAPS)</label>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="form-check form-check-inline">
                            <input class="form-check-input hidden" type="checkbox" tabindex="0" name="[ldapCheckCert]:boolean" v-model="parameter.ldap_skip_verify">
                            <label>{{ $t('admin.global.ldap.skipverify') }}</label>
                        </div>
                    </div>
                </div>
                <br>
                <div class="row">
                    <div class="col-md-3">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.server') }}</span>
                            <input type="text" class="form-control" name="[ldapHost]" v-model="parameter.ldap_host">
                        </div>
                    </div>
                    <div class="col-md-2">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.port') }}</span>
                            <input type="number" class="form-control" name="[ldapPort]" placeholder="389 / 636" min="0" v-model="parameter.ldap_port">
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.basedn') }}</span>
                            <input type="text" class="form-control" name="[ldapBaseDN]" v-model="parameter.ldap_base_dn">
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.login') }}</span>
                            <input type="text" class="form-control" name="[ldapFilter]" placeholder="uid, samaccountname, mail, ..." v-model="parameter.ldap_filter">
                        </div>
                    </div>
                </div>
                <div class="row">
                    <div class="col-md-5">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.user') }}</span>
                            <input type="text" class="form-control" name="[ldapUser]" v-model="parameter.ldap_user">
                        </div>
                    </div>
                    <div class="col-md-5">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.ldap.password') }}</span>
                            <input type="password" class="form-control" name="[ldapPassword]" v-model="parameter.ldap_password">
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <br>
        <div class="card">
            <div class="card-header text-center">{{ $t('admin.global.export.header') }}</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <div class="input-group mb-3">
                            <span class="input-group-text">{{ $t('admin.global.export.delimiter') }}</span>
                            <input type="text" class="form-control" v-model="parameter.export_csv_delimiter" placeholder="Example --> ; , |">
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
</template>