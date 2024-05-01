<script setup>
import { ref, onMounted, inject } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');

const Views = ref(null)
const Groups = ref(null)
const Acls = ref(null)
const ViewAction = ref(null)

const fetchViews = async () => {
    axios.get(`${apiUrl}/views`)
    .then(function (response) {
        Views.value = response.data
    })
    .catch(function (error) {
        code.value = error
        console.error(`Erreur lors de la récupération des utilisateurs`, error);
    });
};

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

const fetchAcls = async () => {
    axios.get(`${apiUrl}/acl`)
    .then(function (response) {
        Acls.value = response.data
    })
    .catch(function (error) {
        code.value = error
        console.error(`Erreur lors de la récupération des rôles`, error);
    });
};

const toggleGroup = (groupId) => {
    if (Acls.value.views && Acls.value.views[ViewAction.value.id] && Acls.value.views[ViewAction.value.id].allow_gid.includes(groupId)) {
        RemoveAcl(ViewAction.value.id, groupId);
    } else {
        AddAcl(ViewAction.value.id, groupId);
    }
};

function AddAcl(vid, gid) {
    axios.post(`${apiUrl}/acl`, {
        "view": vid,
        "gid": gid
    })
    .then(function (response) {
        fetchAcls();
        // console.log(response);
    })
    .catch(function (error) {
        // console.log(error);
    });
}

function RemoveAcl(vid, gid) {
    axios.delete(`${apiUrl}/acl/view/${vid}/group/${gid}`)
    .then(function (response) {
        fetchAcls();
        // console.log(response);
    })
    .catch(function (error) {
        // console.log(error);
    });
}

function ToggleProtection() {
    axios.post(`${apiUrl}/view`, {...ViewAction.value, protected: !ViewAction.value.protected})
    .then(function (response) {
        ViewAction.value.protected = !ViewAction.value.protected
        // console.log(response);
    })
    .catch(function (error) {
        // console.log(error);
    });
}

onMounted(async () => {
    await fetchAcls();
    await fetchGroups();
    await fetchViews();
});
</script>

<template>
    <table class="table">
        <thead>
            <tr>
                <th scope="col" class="col-1">#</th>
                <th scope="col" class="col-2">View</th>
                <th scope="col" class="col-6">Groups Allow</th>
                <th scope="col" class="col-2">Actions</th>
            </tr>
        </thead>
        <tbody class="table-group-divider">
            <tr v-for="view in Views" :key="view.id">
                <th scope="row"><span :class="[view.protected ? 'badge text-bg-success' : 'badge text-bg-secondary']" :title="view.protected ? 'View is readable only by authorized groups' : 'View is readable by everyone'"><i :class="[view.protected ? 'bi bi-lock-fill' : 'bi bi-unlock']"></i></span> {{ view.id }}</th>
                <td> {{ view.name }} </td>
                <td>
                    <template v-if="Acls && Acls.views && Acls.views[view.id] && Groups" v-for="gid in Acls.views[view.id]['allow_gid']">
                        <span v-if="gid == 1" class="badge text-bg-primary me-1">{{ Groups[gid].name }}</span>
                        <span v-if="gid != 1" class="badge text-bg-success me-1">{{ Groups[gid].name }}</span>
                    </template>
                </td>
                <td>
                    <button type="button" class="btn btn-outline-primary btn-sm me-1" title="Toogle protection" @click="ViewAction = view; ToggleProtection()"><i :class="[view.protected ? 'bi bi-unlock' : 'bi bi-lock-fill']"></i></button>
                    <button type="button" class="btn btn-outline-success btn-sm me-1" title="Allow groups" @click="ViewAction = view" data-bs-toggle="modal" data-bs-target="#ChangeGroups"><i class="bi bi-people-fill"></i></button>
                </td>
            </tr>
        </tbody>
    </table>

    <div class="modal fade" id="ChangeGroups" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div v-if="ViewAction" class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">View : {{ ViewAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div v-if="ViewAction.name" class="mt-3">
                        <h5>Groupes :</h5>
                        <div v-if="Groups" v-for="group in Groups" :key="group.id" class="form-check">
                            <input class="form-check-input" type="checkbox" :id="'groupCheckbox_' + group.id" :checked="Acls.views && Acls.views[ViewAction.id] && Acls.views[ViewAction.id].allow_gid.includes(group.id)" @change="toggleGroup(group.id)">
                            <label class="form-check-label" :for="'groupCheckbox_' + group.id">{{ group.name }}</label>
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" @click="" data-bs-dismiss="modal">Ok</button>
                </div>
            </div>
        </div>
    </div>
</template>