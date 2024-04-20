<script setup>
import { ref, onMounted, inject, watch, reactive } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');

const Groups = ref(null)

const GroupAction = ref({
    id: null,
    name: null,
    description: null
})

const NewGroup = ref({
    id: 0,
    name: null,
    description: null
})


const fetchGroups = async () => {
    axios.get(`${apiUrl}/groups`)
    .then(function (response) {
        Groups.value = response.data
    })
    .catch(function (error) {
        code.value = error
        console.error(`Erreur lors de la récupération des groupes`, error);
    });
};

function AddGroup() {
    console.log(NewGroup.value)
    axios.post(`${apiUrl}/group`, NewGroup.value)
    .then(function (response) {
        NewGroup.value = {
                            id: 0,
                            name: null,
                            description: null
                        }
        fetchGroups();
    })
    .catch(function (error) {
        console.log(error);
    });
}

function UpdateGroup() {
    console.log(GroupAction.value)
    axios.put(`${apiUrl}/group/${GroupAction.value.id}`, GroupAction.value)
    .then(function (response) {
        fetchGroups();
    })
    .catch(function (error) {
        console.log(error);
    });
}

function RemoveGroup(id) {
    axios.delete(`${apiUrl}/group/${id}`)
    .then(function (response) {
        fetchGroups();
    })
}

function checkGroupname(targetName) {
    if (Groups.value) {
        const foundItem = Object.values(Groups.value).filter(item => item.id !== 0).find(item => item.name === targetName);

        return !!foundItem;
    } else {
        return false
    }
}

function OnChangeNewGroupname() {
    if (NewGroup.value.name) {
        // Si le champ n'est pas vide, utilisez l'ID 0 pour le nouvel utilisateur
        Groups.value[0] = NewGroup.value;
    } else {
        // Si le champ est vide, supprimez l'ID 0 s'il existe
        if (Groups.value[0]) {
            delete Groups.value[0];
        }
    }
}

onMounted(() => {
    fetchGroups();
});
</script>

<template>
    <table class="table">
        <thead>
            <tr>
                <th scope="col" class="col-1">#</th>
                <th scope="col" class="col-2">Nom du groupe</th>
                <th scope="col" class="col-4">Description</th>
                <th scope="col" class="col-1">Actions</th>
            </tr>
            <tr>
                <td scope="col" colspan="2">
                    <input type="text" class="form-control form-control-sm" :class="{'is-invalid': checkGroupname(NewGroup.name)}" placeholder="Groupname" aria-label="Groupname" v-model="NewGroup.name" @input="OnChangeNewGroupname()">
                </td>
                <td scope="col">
                    <input type="text" class="form-control form-control-sm" placeholder="Description ..." aria-label="Description" v-model="NewGroup.description">
                </td>
                <td scope="col">
                    <div class="input-group">
                        <button type="button" class="btn btn-outline-success btn-sm" title="Add Group" @click="AddGroup()" :disabled="!NewGroup.name ||checkGroupname(NewGroup.name)"><i class="bi bi-plus"></i></button>
                    </div>
                </td>
            </tr>
        </thead>
        <tbody class="table-group-divider">
            <tr v-for="group in Groups" :key="group.id">
                <th scope="row">{{ group.id }}</th>
                <td>{{ group.name }}</td>
                <td>{{ group.description }}</td>
                <td>
                    <template v-if="group.id != 0" >
                        <button v-if="group.id != 1" type="button" class="btn btn-outline-primary btn-sm me-1" title="Change password" @click="GroupAction = Object.assign({}, group)" data-bs-toggle="modal" data-bs-target="#UpdateGroup"><i class="bi bi-pencil-square"></i></button>
                        <button v-if="group.name != 'admin'" type="button" class="btn btn-outline-danger btn-sm" title="Delete Group" @click="GroupAction = group" data-bs-toggle="modal" data-bs-target="#DeleteGroup"><i class="bi bi-trash2"></i></button>
                    </template>
                </td>
            </tr>
        </tbody>
    </table>


    <div class="modal fade" id="DeleteGroup" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">Group : {{ GroupAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p>Etes vous sûr de vouloir supprimer le groupe {{ GroupAction.name }} ?</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
                    <button type="button" class="btn btn-danger" @click="RemoveGroup(GroupAction.id)" data-bs-dismiss="modal">Oui</button>
                </div>
            </div>
        </div>
    </div>


    <div class="modal fade" id="UpdateGroup" tabindex="-1" aria-labelledby="UpdateGroupLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="UpdateGroupLabel">Group : #{{ GroupAction.id }} {{ GroupAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <label for="groupName" class="form-label">Nom du groupe</label>
                    <input type="text" class="form-control" id="groupName" v-model="GroupAction.name">
                    <br>
                    <label for="groupDescription" class="form-label">Description</label>
                    <input type="text" class="form-control" id="groupDescription" v-model="GroupAction.description">
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
                    <button type="button" class="btn btn-primary" @click="UpdateGroup()" data-bs-dismiss="modal">Valider</button>
                </div>
            </div>
        </div>
    </div>

</template>
