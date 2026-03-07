<script setup>
import { computed, ref, onMounted, inject } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');
const myUser = inject('myUser');
const i18n = inject('i18n');

const Users = ref({})
const Groups = ref({})
const Roles = ref({})
const UserAction = ref({
    id: null,
    name: null
})

const NewUser = ref({
    id: 0,
    name: null,
    type: 'local',
    password: null,
    groups: []
})

const usersList = computed(() =>
    Object.values(Users.value || {})
        .filter((user) => user && user.id !== undefined)
        .sort((a, b) => a.id - b.id)
);

const groupsList = computed(() =>
    Object.values(Groups.value || {})
        .filter((group) => group && group.id !== undefined)
        .sort((a, b) => a.id - b.id)
);

const totalUsers = computed(() => usersList.value.length);

const isNewUserDuplicate = computed(() => checkUsername(NewUser.value.name));

const canCreateUser = computed(() => {
    const hasUsername = !!NewUser.value.name;
    const hasValidType = NewUser.value.type === 'ldap' || NewUser.value.type === 'local';
    const hasPasswordIfLocal = NewUser.value.type === 'ldap' || !!NewUser.value.password;
    return hasUsername && hasValidType && hasPasswordIfLocal && !isNewUserDuplicate.value;
});

const fetchUsers = async () => {
    try {
        const response = await axios.get(`${apiUrl}/users`);
        Users.value = response.data || {};
    } catch (error) {
        console.error('Failed to fetch users', error);
    }
};

const fetchGroups = async () => {
    try {
        const response = await axios.get(`${apiUrl}/groups`);
        Groups.value = response.data || {};
    } catch (error) {
        console.error('Failed to fetch groups', error);
    }
};

const fetchUsersRoles = async () => {
    try {
        const response = await axios.get(`${apiUrl}/roles/users`);
        Roles.value = response.data || {};
    } catch (error) {
        console.error('Failed to fetch roles', error);
    }
};

const getGroupById = (groupId) => {
    if (!Groups.value) {
        return null;
    }
    return Groups.value[groupId] || groupsList.value.find((group) => group.id === groupId) || null;
};

const getGroupName = (groupId) => {
    const group = getGroupById(groupId);
    return group?.name || `#${groupId}`;
};

const selectedNewUserGroups = computed(() =>
    (NewUser.value.groups || []).map((groupId) => ({
        id: groupId,
        name: getGroupName(groupId),
    }))
);

const openUpdateUser = (user) => {
    UserAction.value = {
        ...user,
        password: null,
    };
};

const openDeleteUser = (user) => {
    UserAction.value = user;
};

const openGroupModal = (user) => {
    UserAction.value = user;
};

const openNewUserGroupModal = () => {
    UserAction.value = NewUser.value;
};

const isGroupChecked = (groupId) => {
    if (!UserAction.value) {
        return false;
    }

    if (UserAction.value.id === 0) {
        return (NewUser.value.groups || []).includes(groupId);
    }

    return Roles.value?.[UserAction.value.id]?.includes(groupId) || false;
};

const toggleGroup = (groupId) => {
    if (!UserAction.value) {
        return;
    }

    if (UserAction.value.id === 0) {
        if (NewUser.value.groups.includes(groupId)) {
            NewUser.value.groups = NewUser.value.groups.filter((id) => id !== groupId);
        } else {
            NewUser.value.groups.push(groupId);
        }
        return;
    }

    if (Roles.value[UserAction.value.id] && Roles.value[UserAction.value.id].includes(groupId)) {
        RoleRemove(UserAction.value.id, groupId);
        Roles.value[UserAction.value.id] = Roles.value[UserAction.value.id].filter((id) => id !== groupId);
    } else {
        RoleAdd(UserAction.value.id, groupId);
        if (!Roles.value[UserAction.value.id]) {
            Roles.value[UserAction.value.id] = [];
        }
        Roles.value[UserAction.value.id].push(groupId);
    }
};

const resetNewUserForm = () => {
    NewUser.value = {
        id: 0,
        name: null,
        type: 'local',
        password: null,
        groups: []
    };
};

async function AddUser() {
    try {
        const response = await axios.post(`${apiUrl}/user`, NewUser.value);
        const createdUserId = response?.data?.id;

        if (createdUserId && NewUser.value.groups.length > 0) {
            await Promise.all(
                NewUser.value.groups.map((groupId) => RoleAdd(createdUserId, groupId))
            );
        }

        resetNewUserForm();
        await Promise.all([fetchUsers(), fetchUsersRoles()]);
    } catch (error) {
        console.log(error);
    }
}

async function UserUpdate() {
    try {
        await axios.put(`${apiUrl}/user/${UserAction.value.id}`, UserAction.value);
        await fetchUsers();
    } catch (error) {
        console.log(error);
    }
}

async function RemoveUser(id) {
    try {
        await axios.delete(`${apiUrl}/user/${id}`);
        await fetchUsers();
    } catch (error) {
        console.log(error);
    }
}
function RoleAdd(uid, gid) {
    return axios.post(`${apiUrl}/roles`, {
        "user": uid,
        "group": gid
    })
        .catch(function (error) {
            console.log(error);
        });
}

function RoleRemove(uid, gid) {
    return axios.delete(`${apiUrl}/roles/user/${uid}/group/${gid}`)
        .catch(function (error) {
            console.log(error);
        });
}

function checkUsername(targetName) {
    if (!targetName || !Users.value) {
        return false;
    }

    if (Users.value) {
        const foundItem = Object.values(Users.value)
            .filter((item) => item.id !== 0)
            .find((item) => item.name === targetName);

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
    <section class="admin-users-page container-fluid px-0 py-1 py-lg-2">
        <div class="d-flex flex-column gap-3 gap-xxl-4">
            <header class="card admin-users-hero shadow-sm">
                <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
                    <div class="admin-users-hero-icon">
                        <i class="bi bi-people-fill"></i>
                    </div>
                    <div class="flex-grow-1">
                        <p class="admin-users-kicker mb-1">{{ $t('admin.header') }}</p>
                        <h4 class="mb-1">{{ $t('admin.users.header') }}</h4>
                        <p class="mb-0 text-secondary">{{ $t('admin.users.subtitle') }}</p>
                    </div>
                    <span class="badge rounded-pill admin-users-state-chip text-bg-info">
                        <i class="bi bi-person-badge-fill me-1"></i>
                        {{ $t('admin.users.total') }}: {{ totalUsers }}
                    </span>
                </div>
            </header>

            <div class="row g-3 g-xxl-4">
                <div class="col-12 col-xl-4">
                    <article class="card admin-users-panel shadow-sm h-100">
                        <div class="card-body p-4 d-flex flex-column gap-3">
                            <h5 class="admin-users-panel-title mb-0">{{ $t('admin.users.adduser') }}</h5>

                            <div>
                                <label class="form-label">{{ $t('global.username') }}</label>
                                <input type="text" class="form-control" :class="{ 'is-invalid': isNewUserDuplicate }"
                                    :placeholder="$t('global.username')" :aria-label="$t('global.username')"
                                    v-model="NewUser.name">
                                <div v-if="isNewUserDuplicate" class="invalid-feedback">
                                    {{ $t('admin.users.usernameexists') }}
                                </div>
                            </div>

                            <div>
                                <label class="form-label">{{ $t('admin.users.type') }}</label>
                                <select class="form-select" v-model="NewUser.type">
                                    <option value="local">{{ $t('admin.users.local') }}</option>
                                    <option value="ldap">{{ $t('admin.users.ldap') }}</option>
                                </select>
                            </div>

                            <div v-if="NewUser.type === 'local'">
                                <label class="form-label">{{ $t('global.password') }}</label>
                                <input type="password" class="form-control" :placeholder="$t('global.password')"
                                    :aria-label="$t('global.password')" v-model="NewUser.password">
                            </div>

                            <div>
                                <div class="d-flex align-items-center justify-content-between mb-2">
                                    <label class="form-label mb-0">{{ $t('admin.users.groups') }}</label>
                                    <button type="button" class="btn btn-outline-secondary btn-sm"
                                        :title="i18n.global.t('admin.users.changegroups')" :disabled="!NewUser.name"
                                        @click="openNewUserGroupModal" data-bs-toggle="modal"
                                        data-bs-target="#ChangeGroups">
                                        <i class="bi bi-people-fill me-1"></i>
                                        {{ $t('admin.users.changegroups') }}
                                    </button>
                                </div>
                                <div class="admin-users-group-preview">
                                    <span v-if="selectedNewUserGroups.length === 0" class="text-secondary small">
                                        {{ $t('admin.users.nogroup') }}
                                    </span>
                                    <span v-for="group in selectedNewUserGroups" :key="group.id"
                                        class="badge rounded-pill text-bg-secondary me-1 mb-1">
                                        {{ group.name }}
                                    </span>
                                </div>
                            </div>

                            <button type="button"
                                class="btn btn-success d-inline-flex align-items-center gap-2 align-self-start"
                                :title="i18n.global.t('admin.users.adduser')" :disabled="!canCreateUser"
                                @click="AddUser()">
                                <i class="bi bi-person-plus-fill"></i>
                                <span>{{ $t('admin.users.adduser') }}</span>
                            </button>
                        </div>
                    </article>
                </div>

                <div class="col-12 col-xl-8">
                    <article class="card admin-users-panel admin-users-table-panel shadow-sm h-100">
                        <div class="card-body p-0 d-flex flex-column">
                            <div class="admin-users-panel-head px-3 px-lg-4 py-3">
                                <h5 class="admin-users-panel-title mb-1">{{ $t('admin.users.header') }}</h5>
                                <p class="small text-secondary mb-0">{{ $t('admin.users.tablehint') }}</p>
                            </div>

                            <div class="table-responsive admin-users-table-wrap">
                                <table class="table align-middle mb-0 admin-users-table">
                                    <thead>
                                        <tr>
                                            <th scope="col" class="col-1">#</th>
                                            <th scope="col" class="col-3">{{ $t('global.username') }}</th>
                                            <th scope="col" class="col-2">{{ $t('admin.users.type') }}</th>
                                            <th scope="col" class="col-4">{{ $t('admin.users.groups') }}</th>
                                            <th scope="col" class="col-2 text-end">{{ $t('admin.users.actions') }}</th>
                                        </tr>
                                    </thead>
                                    <tbody class="table-group-divider">
                                        <tr v-if="usersList.length === 0">
                                            <td colspan="5" class="text-center text-secondary py-4">-</td>
                                        </tr>
                                        <tr v-for="user in usersList" :key="user.id">
                                            <th scope="row">{{ user.id }}</th>
                                            <td>{{ user.name }}</td>
                                            <td>
                                                <span class="badge rounded-pill"
                                                    :class="user.type === 'local' ? 'text-bg-primary' : 'text-bg-secondary'">
                                                    {{ user.type === 'local' ? $t('admin.users.local') :
                                                        $t('admin.users.ldap') }}
                                                </span>
                                            </td>
                                            <td>
                                                <template v-if="Roles && Roles[user.id] && Roles[user.id].length">
                                                    <span v-for="role in Roles[user.id]" :key="role"
                                                        class="badge rounded-pill me-1 mb-1"
                                                        :class="role === 1 ? 'text-bg-primary' : 'text-bg-secondary'">
                                                        {{ getGroupName(role) }}
                                                    </span>
                                                </template>
                                                <span v-else class="text-secondary">-</span>
                                            </td>
                                            <td class="text-end">
                                                <button type="button" class="btn btn-outline-primary btn-sm me-1"
                                                    :title="i18n.global.t('admin.users.edituser')"
                                                    @click="openUpdateUser(user)" data-bs-toggle="modal"
                                                    data-bs-target="#UpdateUser">
                                                    <i class="bi bi-pencil-square"></i>
                                                </button>

                                                <button v-if="user.name !== 'admin'" type="button"
                                                    class="btn btn-outline-success btn-sm me-1"
                                                    :title="i18n.global.t('admin.users.changegroups')"
                                                    @click="openGroupModal(user)" data-bs-toggle="modal"
                                                    data-bs-target="#ChangeGroups">
                                                    <i class="bi bi-people-fill"></i>
                                                </button>

                                                <button v-if="user.name !== 'admin'" type="button"
                                                    class="btn btn-outline-danger btn-sm"
                                                    :title="i18n.global.t('admin.users.deleteuser')"
                                                    @click="openDeleteUser(user)" data-bs-toggle="modal"
                                                    data-bs-target="#DeleteUser">
                                                    <i class="bi bi-trash2"></i>
                                                </button>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </article>
                </div>
            </div>
        </div>
    </section>

    <div v-if="UserAction" class="modal fade" id="UpdateUser" tabindex="-1" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('admin.users.user') }} : {{
                        UserAction.name }}
                    </h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div v-if="UserAction.id != 1" class="modal-body">
                    <div class="mb-3">
                        <label for="InputName" class="form-label">{{ $t('global.username') }}</label>
                        <input type="text" class="form-control" id="InputName" v-model="UserAction.name">
                    </div>
                    <div class="mb-3">
                        <label for="InputType" class="form-label">{{ $t('admin.users.type') }}</label>
                        <select class="form-select" id="InputType" v-model="UserAction.type">
                            <option value="local">{{ $t('admin.users.local') }}</option>
                            <option value="ldap">{{ $t('admin.users.ldap') }}</option>
                        </select>
                    </div>
                </div>
                <div v-if="UserAction.type == 'local'" class="modal-body">
                    <div class="mb-3">
                        <label for="InputName" class="form-label">{{ $t('global.password') }}</label>
                        <input type="password" class="form-control" id="InputName" v-model="UserAction.password">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel')
                    }}</button>
                    <button type="button" class="btn btn-primary" @click="UserUpdate()" data-bs-dismiss="modal">{{
                        $t('global.apply') }}</button>
                </div>
            </div>
        </div>
    </div>

    <div v-if="UserAction" class="modal fade" id="DeleteUser" tabindex="-1" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('admin.users.user') }} : {{
                        UserAction.name }}
                    </h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p>{{ $t('admin.users.askdelete') }} {{ UserAction.name }} ?</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel')
                    }}</button>
                    <button type="button" class="btn btn-danger" @click="RemoveUser(UserAction.id)"
                        data-bs-dismiss="modal">{{ $t('global.yes') }}</button>
                </div>
            </div>
        </div>
    </div>

    <div v-if="UserAction" class="modal fade" id="ChangeGroups" tabindex="-1" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('global.username') }} : {{ UserAction.name
                    }}
                    </h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div v-if="UserAction.name" class="mt-3">
                        <h5>{{ $t('global.groups') }} :</h5>
                        <div v-for="group in groupsList" :key="group.id" class="form-check">
                            <input class="form-check-input" type="checkbox" :id="'groupCheckbox_' + group.id"
                                :checked="isGroupChecked(group.id)" :disabled="UserAction.id == myUser.id"
                                @change="toggleGroup(group.id)">
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