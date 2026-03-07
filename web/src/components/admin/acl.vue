<script setup>
import { computed, ref, onMounted, inject } from "vue";
import axios from 'axios';

const apiUrl = inject('apiUrl');

const Views = ref({})
const Groups = ref({})
const Acls = ref({ views: {} })
const ViewAction = ref(null)

const viewsList = computed(() =>
    Object.values(Views.value || {})
        .filter((view) => view && view.id !== undefined)
        .sort((a, b) => a.id - b.id)
);

const groupsList = computed(() =>
    Object.values(Groups.value || {})
        .filter((group) => group && group.id !== undefined)
        .sort((a, b) => a.id - b.id)
);

const totalViews = computed(() => viewsList.value.length);
const protectedViews = computed(() => viewsList.value.filter((view) => view.protected).length);

const getAclGroupIds = (viewId) => {
    const aclView = Acls.value?.views?.[viewId];
    if (!aclView || !Array.isArray(aclView.allow_gid)) {
        return [];
    }
    return aclView.allow_gid;
};

const getGroupName = (groupId) => {
    const byMap = Groups.value?.[groupId];
    if (byMap?.name) {
        return byMap.name;
    }

    const byList = groupsList.value.find((group) => group.id === groupId);
    return byList?.name || `#${groupId}`;
};

const isGroupAllowed = (viewId, groupId) => getAclGroupIds(viewId).includes(groupId);

const selectedViewGroups = computed(() => {
    if (!ViewAction.value) {
        return [];
    }
    return getAclGroupIds(ViewAction.value.id).map((groupId) => ({
        id: groupId,
        name: getGroupName(groupId),
    }));
});

const fetchViews = async () => {
    try {
        const response = await axios.get(`${apiUrl}/views`);
        Views.value = response.data || {};
    } catch (error) {
        console.error('Failed to fetch views', error);
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

const fetchAcls = async () => {
    try {
        const response = await axios.get(`${apiUrl}/acl`);
        Acls.value = response.data || { views: {} };
    } catch (error) {
        console.error('Failed to fetch ACL entries', error);
    }
};

const openAclModal = (view) => {
    ViewAction.value = view;
};

const toggleGroup = async (groupId) => {
    if (!ViewAction.value) {
        return;
    }

    const viewId = ViewAction.value.id;
    if (isGroupAllowed(viewId, groupId)) {
        await RemoveAcl(viewId, groupId);
    } else {
        await AddAcl(viewId, groupId);
    }
};

async function AddAcl(vid, gid) {
    try {
        await axios.post(`${apiUrl}/acl`, {
            "view": vid,
            "gid": gid
        });
        await fetchAcls();
    } catch (error) {
        console.error('Failed to add ACL entry', error);
    }
}

async function RemoveAcl(vid, gid) {
    try {
        await axios.delete(`${apiUrl}/acl/view/${vid}/group/${gid}`);
        await fetchAcls();
    } catch (error) {
        console.error('Failed to remove ACL entry', error);
    }
}

async function ToggleProtection(view) {
    if (!view) {
        return;
    }

    const nextProtectedState = !view.protected;
    try {
        await axios.post(`${apiUrl}/view`, {
            ...view,
            protected: nextProtectedState,
        });
        view.protected = nextProtectedState;
    } catch (error) {
        console.error('Failed to toggle protection', error);
    }
}

onMounted(async () => {
    await Promise.all([fetchAcls(), fetchGroups(), fetchViews()]);
});
</script>

<template>
    <section class="admin-acl-page container-fluid px-0 py-1 py-lg-2">
        <div class="d-flex flex-column gap-3 gap-xxl-4">
            <header class="card admin-acl-hero shadow-sm">
                <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
                    <div class="admin-acl-hero-icon">
                        <i class="bi bi-shield-lock-fill"></i>
                    </div>
                    <div class="flex-grow-1">
                        <p class="admin-acl-kicker mb-1">{{ $t('admin.header') }}</p>
                        <h4 class="mb-1">{{ $t('admin.acl.header') }}</h4>
                        <p class="mb-0 text-secondary">{{ $t('admin.acl.subtitle') }}</p>
                    </div>
                    <div class="d-flex gap-2 flex-wrap justify-content-lg-end">
                        <span class="badge rounded-pill admin-acl-state-chip text-bg-info">
                            <i class="bi bi-grid-3x3-gap-fill me-1"></i>
                            {{ $t('admin.acl.total') }}: {{ totalViews }}
                        </span>
                        <span class="badge rounded-pill admin-acl-state-chip text-bg-success">
                            <i class="bi bi-shield-lock-fill me-1"></i>
                            {{ $t('admin.acl.protectedcount') }}: {{ protectedViews }}
                        </span>
                    </div>
                </div>
            </header>

            <article class="card admin-acl-panel admin-acl-table-panel shadow-sm">
                <div class="card-body p-0 d-flex flex-column">
                    <div class="admin-acl-panel-head px-3 px-lg-4 py-3">
                        <h5 class="admin-acl-panel-title mb-1">{{ $t('admin.acl.header') }}</h5>
                        <p class="small text-secondary mb-0">{{ $t('admin.acl.tablehint') }}</p>
                    </div>

                    <div class="table-responsive admin-acl-table-wrap">
                        <table class="table align-middle mb-0 admin-acl-table">
                            <thead>
                                <tr>
                                    <th scope="col" class="col-1">#</th>
                                    <th scope="col" class="col-3">{{ $t('admin.acl.view') }}</th>
                                    <th scope="col" class="col-2">{{ $t('admin.acl.status') }}</th>
                                    <th scope="col" class="col-4">{{ $t('admin.acl.groupsallow') }}</th>
                                    <th scope="col" class="col-2 text-end">{{ $t('admin.acl.actions') }}</th>
                                </tr>
                            </thead>
                            <tbody class="table-group-divider">
                                <tr v-if="viewsList.length === 0">
                                    <td colspan="5" class="text-center text-secondary py-4">-</td>
                                </tr>
                                <tr v-for="view in viewsList" :key="view.id">
                                    <th scope="row">{{ view.id }}</th>
                                    <td>{{ view.name }}</td>
                                    <td>
                                        <span class="badge rounded-pill"
                                            :class="view.protected ? 'text-bg-success' : 'text-bg-secondary'">
                                            <i
                                                :class="view.protected ? 'bi bi-lock-fill me-1' : 'bi bi-unlock me-1'"></i>
                                            {{ view.protected ? $t('admin.acl.protected') : $t('admin.acl.public') }}
                                        </span>
                                    </td>
                                    <td>
                                        <template v-if="getAclGroupIds(view.id).length">
                                            <span v-for="gid in getAclGroupIds(view.id)" :key="gid"
                                                class="badge rounded-pill me-1 mb-1"
                                                :class="gid === 1 ? 'text-bg-primary' : 'text-bg-success'">
                                                {{ getGroupName(gid) }}
                                            </span>
                                        </template>
                                        <span v-else class="text-secondary">-</span>
                                    </td>
                                    <td class="text-end">
                                        <button type="button" class="btn btn-outline-primary btn-sm me-1"
                                            :title="$t('admin.acl.toggleprotection')" @click="ToggleProtection(view)">
                                            <i :class="view.protected ? 'bi bi-unlock' : 'bi bi-lock-fill'"></i>
                                        </button>
                                        <button type="button" class="btn btn-outline-success btn-sm"
                                            :title="$t('admin.acl.allowgroups')" @click="openAclModal(view)"
                                            data-bs-toggle="modal" data-bs-target="#ChangeGroups">
                                            <i class="bi bi-people-fill"></i>
                                        </button>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </article>
        </div>
    </section>

    <div class="modal fade" id="ChangeGroups" tabindex="-1" aria-labelledby="aclGroupsLabel" aria-hidden="true">
        <div class="modal-dialog">
            <div v-if="ViewAction" class="modal-content">
                <div class="modal-header">
                    <h1 class="modal-title fs-5" id="aclGroupsLabel">{{ $t('admin.acl.view') }} : {{ ViewAction.name }}
                    </h1>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="mb-3 d-flex flex-wrap gap-2">
                        <span class="badge rounded-pill"
                            :class="ViewAction.protected ? 'text-bg-success' : 'text-bg-secondary'">
                            {{ ViewAction.protected ? $t('admin.acl.protected') : $t('admin.acl.public') }}
                        </span>
                        <span v-for="group in selectedViewGroups" :key="group.id"
                            class="badge rounded-pill text-bg-secondary">
                            {{ group.name }}
                        </span>
                    </div>

                    <div v-if="groupsList.length" class="mt-2">
                        <h6 class="mb-3">{{ $t('admin.acl.groupsallow') }}</h6>
                        <div v-for="group in groupsList" :key="group.id" class="form-check mb-2">
                            <input class="form-check-input" type="checkbox" :id="'groupCheckbox_' + group.id"
                                :checked="isGroupAllowed(ViewAction.id, group.id)" @change="toggleGroup(group.id)">
                            <label class="form-check-label" :for="'groupCheckbox_' + group.id">{{ group.name }}</label>
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel')
                        }}</button>
                </div>
            </div>
        </div>
    </div>
</template>