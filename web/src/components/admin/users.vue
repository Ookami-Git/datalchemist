<script setup>
import { ref, onMounted, inject } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');
const myUser = inject('myUser');

const Users = ref(null)
const Groups = ref(null)
const Roles = ref(null)
const UserAction = ref({
    id: null,
    name: null
})

const NewUser = ref({
    id: 0,
    name: null,
    type: null,
    password: null,
    groups: []
})

const fetchUsers = async () => {
    axios.get(`${apiUrl}/users`)
    .then(function (response) {
        Users.value = response.data
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

const fetchUsersRoles = async () => {
    axios.get(`${apiUrl}/roles/users`)
    .then(function (response) {
        //console.log(response.data)
        Roles.value = response.data
    })
    .catch(function (error) {
        code.value = error
        console.error(`Erreur lors de la récupération des rôles`, error);
    });
};

const toggleGroup = (groupId) => {
    if (Roles.value[UserAction.value.id] && Roles.value[UserAction.value.id].includes(groupId)) {
        if (UserAction.value.id != 0) {
            RoleRemove(UserAction.value.id, groupId);
        }
        Roles.value[UserAction.value.id] = Roles.value[UserAction.value.id].filter(id => id !== groupId);
    } else {
        if (UserAction.value.id != 0) {
            RoleAdd(UserAction.value.id, groupId);
        }
        if (!Roles.value[UserAction.value.id]) {
            Roles.value[UserAction.value.id] = [];
        }
        Roles.value[UserAction.value.id].push(groupId);
    }
};

function OnChangeNewUsername() {
    if (NewUser.value.name) {
        // Si le champ n'est pas vide, utilisez l'ID 0 pour le nouvel utilisateur
        Users.value[0] = NewUser.value;
        Roles.value[0] = NewUser.value.groups;
    } else {
        // Si le champ est vide, supprimez l'ID 0 s'il existe
        if (Users.value[0]) {
            delete Users.value[0];
            delete Roles.value[0];
        }
    }
}

function AddUser() {
    axios.post(`${apiUrl}/user`, NewUser.value)
    .then(function (response) {
        for (let i = 0; i < NewUser.value.groups.length; i++) {
            RoleAdd(response.data.id, NewUser.value.groups[i]);
        }
        NewUser.value = {
                            id: 0,
                            name: null,
                            type: null,
                            password: null,
                            groups: []
                        }

        fetchUsers();
        fetchUsersRoles();
    })
    .catch(function (error) {
        console.log(error);
    });
}

function UserUpdate() {
    console.log(UserAction.value)
    axios.put(`${apiUrl}/user/${UserAction.value.id}`, UserAction.value)
    .then(function (response) {
        fetchUsers();
    })
    .catch(function (error) {
        console.log(error);
    });
}

function RemoveUser(id) {
    axios.delete(`${apiUrl}/user/${id}`)
    .then(function (response) {
        fetchUsers();
    })
}
function RoleAdd(uid, gid) {
    axios.post(`${apiUrl}/roles`, {
        "user": uid,
        "group": gid
    })
    .then(function (response) {
        //console.log(response);
    })
    .catch(function (error) {
        console.log(error);
    });
}

function RoleRemove(uid, gid) {
    axios.delete(`${apiUrl}/roles/user/${uid}/group/${gid}`)
    .then(function (response) {
        //console.log(response);
    })
    .catch(function (error) {
        console.log(error);
    });
}
function checkUsername(targetName) {
    if (Users.value) {
        const foundItem = Object.values(Users.value).filter(item => item.id !== 0).find(item => item.name === targetName);

        return !!foundItem;
    } else {
        return false
    }
}

onMounted(() => {
    fetchUsersRoles();
    fetchGroups();
    fetchUsers();
});
</script>

<template>
    <table class="table">
        <thead>
            <tr>
                <th scope="col" class="col-1">#</th>
                <th scope="col" class="col-3">Username</th>
                <th scope="col" class="col-1">Type</th>
                <th scope="col" class="col-5">Groups</th>
                <th scope="col" class="col-2">Actions</th>
            </tr>
            <tr>
                <td scope="col" colspan="2">
                    <input type="text" class="form-control form-control-sm" :class="{'is-invalid': checkUsername(NewUser.name)}" placeholder="Username" aria-label="Username" v-model="NewUser.name" @input="OnChangeNewUsername()">
                </td>
                <td scope="col">
                    <select class="form-select form-select-sm" aria-label="Default select example" v-model="NewUser.type">
                        <option value="local" >Local</option>
                        <option value="ldap">LDAP</option>
                    </select>
                </td>
                <td scope="col"><button type="button" class="btn btn-outline-success btn-sm me-1" title="Change user groups" @click="UserAction = NewUser" data-bs-toggle="modal" data-bs-target="#ChangeGroups" :disabled="!NewUser.name"><i class="bi bi-people-fill"></i></button></td>
                <td scope="col">
                    <div class="input-group">
                        <button type="button" class="btn btn-outline-success btn-sm" title="Add User" @click="AddUser()" :disabled="!NewUser.name || !((NewUser.type == 'local' && NewUser.password) || NewUser.type == 'ldap') || checkUsername(NewUser.name)"><i class="bi bi-person-plus-fill"></i></button>
                        <input v-if="NewUser.type == 'local'" type="password" class="form-control form-control-sm" placeholder="Password" aria-label="Password" :disabled="NewUser.type != 'local'" v-model="NewUser.password">
                    </div>
                </td>
            </tr>
        </thead>
        <tbody class="table-group-divider">
            <tr v-if="Users" v-for="user in Users" :key="user.id">
                <th scope="row">{{ user.id }}</th>
                <td>{{ user.name }}</td>
                <td>{{ user.type }}</td>
                <td>
                    <template v-if="Roles && Roles[user.id] && Groups" v-for="role in Roles[user.id]">
                        <span v-if="role == 1" class="badge text-bg-primary me-1">{{ Groups[role].name }}</span>
                        <span v-if="role != 1" class="badge text-bg-secondary me-1">{{ Groups[role].name }}</span>
                    </template>
                </td>
                <td>
                    <template v-if="user.id != 0" >
                        <button type="button" class="btn btn-outline-primary btn-sm me-1" title="Edit User" @click="UserAction = Object.assign({}, user)" data-bs-toggle="modal" data-bs-target="#UpdateUser"><i class="bi bi-pencil-square"></i></button>
                        <button v-if="user.name != 'admin'" type="button" class="btn btn-outline-success btn-sm me-1" title="Change user groups" @click="UserAction = user" data-bs-toggle="modal" data-bs-target="#ChangeGroups"><i class="bi bi-people-fill"></i></button>
                        <button v-if="user.name != 'admin'" type="button" class="btn btn-outline-danger btn-sm" title="Delete user" @click="UserAction = user" data-bs-toggle="modal" data-bs-target="#DeleteUser"><i class="bi bi-trash2"></i></button>
                    </template>
                    <span v-if="user.id == 0">Preview - New user</span>
                </td>
            </tr>
        </tbody>
    </table>

    <div v-if="UserAction" class="modal fade" id="UpdateUser" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">User : {{ UserAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div v-if="UserAction.id != 1" class="modal-body">
                    <div class="mb-3">
                        <label for="InputName" class="form-label">New name</label>
                        <input type="text" class="form-control" id="InputName" v-model="UserAction.name">
                    </div>
                    <div class="mb-3">
                        <label for="InputType" class="form-label">Change authentication type</label>
                        <select class="form-select" id="InputType" v-model="UserAction.type">
                            <option value="local">local</option>
                            <option value="ldap">ldap</option>
                        </select>
                    </div>
                </div>
                <div v-if="UserAction.type == 'local'" class="modal-body">
                    <div class="mb-3">
                        <label for="InputName" class="form-label">New password</label>
                        <input type="password" class="form-control" id="InputName" v-model="UserAction.password">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
                    <button type="button" class="btn btn-primary" @click="UserUpdate()" data-bs-dismiss="modal">Appliquer</button>
                </div>
            </div>
        </div>
    </div>

    <div v-if="UserAction" class="modal fade" id="DeleteUser" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">User : {{ UserAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p>Etes vous sûr de vouloir supprimer l'utilisateur {{ UserAction.name }} ?</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
                    <button type="button" class="btn btn-danger" @click="RemoveUser(UserAction.id)" data-bs-dismiss="modal">Oui</button>
                </div>
            </div>
        </div>
    </div>

    <div v-if="UserAction" class="modal fade" id="ChangeGroups" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">User : {{ UserAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div v-if="UserAction.name" class="mt-3">
                        <h5>Groupes :</h5>
                        <div v-for="group in Groups" :key="group.id" class="form-check">
                            <input class="form-check-input" type="checkbox" :id="'groupCheckbox_' + group.id" :checked="Roles[UserAction.id] ? Roles[UserAction.id].includes(group.id) : false" :disabled="UserAction.id == myUser.id" @change="toggleGroup(group.id)">
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