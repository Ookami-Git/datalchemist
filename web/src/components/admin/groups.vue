<script setup>
import { computed, ref, onMounted, inject } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');

const Groups = ref({})

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

const groupsList = computed(() =>
    Object.values(Groups.value || {})
        .filter((group) => group && group.id !== undefined)
        .sort((a, b) => a.id - b.id)
);

const totalGroups = computed(() => groupsList.value.length);

const isNewGroupDuplicate = computed(() => checkGroupname(NewGroup.value.name));

const canCreateGroup = computed(() => !!NewGroup.value.name && !isNewGroupDuplicate.value);

const fetchGroups = async () => {
    try {
        const response = await axios.get(`${apiUrl}/groups`);
        Groups.value = response.data || {};
    } catch (error) {
        console.error('Failed to fetch groups', error);
    }
};

function resetNewGroupForm() {
    NewGroup.value = {
        id: 0,
        name: null,
        description: null
    };
}

function openUpdateGroup(group) {
    GroupAction.value = Object.assign({}, group);
}

function openDeleteGroup(group) {
    GroupAction.value = group;
}

async function AddGroup() {
    try {
        await axios.post(`${apiUrl}/group`, NewGroup.value);
        resetNewGroupForm();
        await fetchGroups();
    } catch (error) {
        console.log(error);
    }
}

async function UpdateGroup() {
    try {
        await axios.put(`${apiUrl}/group/${GroupAction.value.id}`, GroupAction.value);
        await fetchGroups();
    } catch (error) {
        console.log(error);
    }
}

async function RemoveGroup(id) {
    try {
        await axios.delete(`${apiUrl}/group/${id}`);
        await fetchGroups();
    } catch (error) {
        console.log(error);
    }
}

function checkGroupname(targetName) {
    if (!targetName || !Groups.value) {
        return false;
    }

    const foundItem = Object.values(Groups.value)
        .filter((item) => item.id !== 0)
        .find((item) => item.name === targetName);

    return !!foundItem;
}

onMounted(() => {
    fetchGroups();
});
</script>

<template>
    <section class="admin-groups-page container-fluid px-0 py-1 py-lg-2">
        <div class="d-flex flex-column gap-3 gap-xxl-4">
            <header class="card admin-groups-hero shadow-sm">
                <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
                    <div class="admin-groups-hero-icon">
                        <i class="bi bi-collection-fill"></i>
                    </div>
                    <div class="flex-grow-1">
                        <p class="admin-groups-kicker mb-1">{{ $t('admin.header') }}</p>
                        <h4 class="mb-1">{{ $t('admin.groups.header') }}</h4>
                        <p class="mb-0 text-secondary">{{ $t('admin.groups.subtitle') }}</p>
                    </div>
                    <span class="badge rounded-pill admin-groups-state-chip text-bg-info">
                        <i class="bi bi-collection me-1"></i>
                        {{ $t('admin.groups.total') }}: {{ totalGroups }}
                    </span>
                </div>
            </header>

            <div class="row g-3 g-xxl-4">
                <div class="col-12 col-xl-4">
                    <article class="card admin-groups-panel shadow-sm h-100">
                        <div class="card-body p-4 d-flex flex-column gap-3">
                            <h5 class="admin-groups-panel-title mb-0">{{ $t('admin.groups.addgroup') }}</h5>

                            <div>
                                <label class="form-label">{{ $t('admin.groups.name') }}</label>
                                <input type="text" class="form-control" :class="{ 'is-invalid': isNewGroupDuplicate }"
                                    :placeholder="$t('admin.groups.name')" :aria-label="$t('admin.groups.name')"
                                    v-model="NewGroup.name">
                                <div v-if="isNewGroupDuplicate" class="invalid-feedback">
                                    {{ $t('admin.groups.nameexists') }}
                                </div>
                            </div>

                            <div>
                                <label class="form-label">{{ $t('admin.groups.description') }}</label>
                                <input type="text" class="form-control" :placeholder="$t('admin.groups.description')"
                                    :aria-label="$t('admin.groups.description')" v-model="NewGroup.description">
                            </div>

                            <button type="button"
                                class="btn btn-success d-inline-flex align-items-center gap-2 align-self-start"
                                :title="$t('admin.groups.addgroup')" :disabled="!canCreateGroup" @click="AddGroup()">
                                <i class="bi bi-plus-lg"></i>
                                <span>{{ $t('admin.groups.addgroup') }}</span>
                            </button>
                        </div>
                    </article>
                </div>

                <div class="col-12 col-xl-8">
                    <article class="card admin-groups-panel shadow-sm h-100">
                        <div class="card-body p-0 d-flex flex-column">
                            <div class="admin-groups-panel-head px-3 px-lg-4 py-3">
                                <h5 class="admin-groups-panel-title mb-1">{{ $t('admin.groups.header') }}</h5>
                                <p class="small text-secondary mb-0">{{ $t('admin.groups.tablehint') }}</p>
                            </div>

                            <div class="table-responsive">
                                <table class="table align-middle mb-0 admin-groups-table">
                                    <thead>
                                        <tr>
                                            <th scope="col" class="col-1">#</th>
                                            <th scope="col" class="col-3">{{ $t('admin.groups.name') }}</th>
                                            <th scope="col" class="col-6">{{ $t('admin.groups.description') }}</th>
                                            <th scope="col" class="col-2 text-end">{{ $t('admin.users.actions') }}</th>
                                        </tr>
                                    </thead>
                                    <tbody class="table-group-divider">
                                        <tr v-if="groupsList.length === 0">
                                            <td colspan="4" class="text-center text-secondary py-4">-</td>
                                        </tr>
                                        <tr v-for="group in groupsList" :key="group.id">
                                            <th scope="row">{{ group.id }}</th>
                                            <td>{{ group.name }}</td>
                                            <td>{{ group.description || '-' }}</td>
                                            <td class="text-end">
                                                <button v-if="group.id !== 1" type="button"
                                                    class="btn btn-outline-primary btn-sm me-1"
                                                    :title="$t('admin.groups.editgroup')"
                                                    @click="openUpdateGroup(group)" data-bs-toggle="modal"
                                                    data-bs-target="#UpdateGroup">
                                                    <i class="bi bi-pencil-square"></i>
                                                </button>
                                                <button v-if="group.name !== 'admin'" type="button"
                                                    class="btn btn-outline-danger btn-sm"
                                                    :title="$t('admin.groups.deletegroup')"
                                                    @click="openDeleteGroup(group)" data-bs-toggle="modal"
                                                    data-bs-target="#DeleteGroup">
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


    <div v-if="GroupAction" class="modal fade" id="DeleteGroup" tabindex="-1" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('admin.groups.header') }} : {{
                        GroupAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p>{{ $t('admin.groups.askdelete') }} {{ GroupAction.name }} ?</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel')
                    }}</button>
                    <button type="button" class="btn btn-danger" @click="RemoveGroup(GroupAction.id)"
                        data-bs-dismiss="modal">{{ $t('global.yes') }}</button>
                </div>
            </div>
        </div>
    </div>


    <div v-if="GroupAction" class="modal fade" id="UpdateGroup" tabindex="-1" aria-labelledby="UpdateGroupLabel"
        aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="UpdateGroupLabel">{{ $t('admin.groups.header') }} : #{{
                        GroupAction.id }} {{ GroupAction.name }}</h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <label for="groupName" class="form-label">{{ $t('admin.groups.name') }}</label>
                    <input type="text" class="form-control" id="groupName" v-model="GroupAction.name">
                    <br>
                    <label for="groupDescription" class="form-label">{{ $t('admin.groups.description') }}</label>
                    <input type="text" class="form-control" id="groupDescription" v-model="GroupAction.description">
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel')
                    }}</button>
                    <button type="button" class="btn btn-primary" @click="UpdateGroup()" data-bs-dismiss="modal">{{
                        $t('global.apply') }}</button>
                </div>
            </div>
        </div>
    </div>

</template>
