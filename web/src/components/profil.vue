<script setup>
import { ref, onMounted, inject, watch } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');
const parameter = inject('parameters');
const i18n = inject('i18n');

const user = ref(null)
const lang = ref([
    { id: 'default', name: 'Par défaut (Pas de prévisualisation)'},
    { id: 'en', name: 'English'},
    { id: 'fr', name: 'French'},
])
const theme = ref([
    { id: 'default', name: 'Par défaut (Pas de prévisualisation)'},
    { id: 'dark', name: 'Sombre'},
    { id: 'light', name: 'Clair'},
])
const passwordConfirm = ref('')
const errorPassword = ref(false)

const fetchUser = async () => {
    axios.get(`${apiUrl}/user`)
    .then(function (response) {
        user.value = response.data
    })
    .catch(function (error) {
        code.value = error
        console.error(`Erreur lors de la récupération de l'utilisateur`, error);
    });
};

function UpdateUser () {
    axios.put(`${apiUrl}/user/${user.value.id}`, user.value)
    .then(function (response) {
        fetchUser();
        user.value.password = ''
        passwordConfirm.value = ''
    })
    .catch(function (error) {
        console.log(error);
    });
}

function checkPassword() {
    if (user.value.password !== passwordConfirm.value) {
        errorPassword.value = true
        return true
    } else {
        errorPassword.value = false
        return false
    }
}

watch(user, () => {
    if (user.value.theme != "default") {
        parameter.value.theme = user.value.theme
    }
    if (user.value.lang != "default") {
        parameter.value.lang = user.value.lang
    }
}, { deep: true });

watch(parameter, () => {
    localStorage.setItem('reloadparameters', true);
}, { deep: true });

onMounted(() => {
    fetchUser();
});
</script>

<template>
    <div class="container">
        <div class="row" v-if="user">
            <div class="col-md-4">
                <div class="card">
                    <div class="card-header">
                        <h5 class="card-title">Profil</h5>
                    </div>
                    <div class="card-body">
                        <p class="card-text">{{ user.name }}</p>
                    </div>
                    <ul class="list-group list-group-flush">
                        <li class="list-group-item">Langue : {{ user.lang }}</li>
                        <li class="list-group-item">Thème : {{ user.theme }}</li>
                    </ul>
                </div>
            </div>
            <div class="col-md-8">
                <div class="card">
                    <div class="card-header">
                        <h5 class="card-title">Paramètres</h5>
                    </div>
                    <div class="card-body">
                        <div class="mb-3">
                            <label for="inputLang" class="form-label">Langue :</label>
                            <select class="form-select" id="inputLang" v-model="user.lang">
                                <option v-for="l in lang" :key="l.id" :value="l.id">{{ l.name }}</option>
                            </select>
                        </div>
                        <div class="mb-3">
                            <label for="inputTheme" class="form-label">Thème :</label>
                            <select class="form-select" id="inputTheme" v-model="user.theme">
                                <option v-for="t in theme" :key="t.id" :value="t.id">{{ t.name }}</option>
                            </select>
                        </div>
                        <div class="mb-3">
                            <label for="inputPassword" class="form-label">Nouveau mot de passe :</label>
                            <input type="password" class="form-control" id="inputPassword" v-model="user.password" :class="{'is-invalid': checkPassword()}">
                        </div>
                        <div class="mb-3">
                            <label for="inputPasswordConfirm" class="form-label">Confirmation du mot de passe :</label>
                            <input type="password" class="form-control" id="inputPasswordConfirm" v-model="passwordConfirm" :class="{'is-invalid': checkPassword()}">
                        </div>
                    </div>
                    <div class="card-footer">
                        <button type="button" class="btn btn-primary" :disabled="errorPassword" @click="UpdateUser()">Mettre à jour</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>


