<script setup>
import { computed, inject } from 'vue';
import { useRoute } from 'vue-router';
import global from './global.vue'
import acl from './acl.vue'
import users from './users.vue'
import groups from './groups.vue'
import navbar from './navbar.vue'

const parameters = inject('parameters');
const save = inject('save');
const skipNextRouteTransition = inject('skipNextRouteTransition', null);
const route = useRoute();

const goNoTransition = () => {
    if (skipNextRouteTransition) {
        skipNextRouteTransition();
    }
};

const tabs = {
    global,
    acl,
    users,
    groups,
    navbar
}

const navigationItems = [
    { page: 'global', icon: 'bi bi-card-checklist', labelKey: 'admin.global.header' },
    { page: 'navbar', icon: 'bi bi-window', labelKey: 'admin.navbar.header' },
    { page: 'users', icon: 'bi bi-people-fill', labelKey: 'admin.users.header' },
    { page: 'groups', icon: 'bi bi-collection-fill', labelKey: 'admin.groups.header' },
    { page: 'acl', icon: 'bi bi-shield-lock-fill', labelKey: 'admin.acl.header' },
];

const activeTab = computed(() => tabs[route.params.page] || global);

const releaseVersion = computed(() => parameters?.value?.release?.version || 'dev');
const releaseDate = computed(() => {
    const rawDate = parameters?.value?.release?.date;
    if (!rawDate) {
        return 'unknown';
    }
    return rawDate.split('T')[0];
});

const hasUnsavedChanges = () => Boolean(save?.value?.show && !save?.value?.disabled);

const isAdminPageSwitch = (to, from) =>
    to?.name === 'admin' &&
    from?.name === 'admin' &&
    to?.params?.page !== from?.params?.page;

const canLeaveAdminTab = async (to, from) => {
    if (!isAdminPageSwitch(to, from) || !hasUnsavedChanges()) {
        return true;
    }

    if (!save?.value?.confirmLeave) {
        return true;
    }

    return await save.value.confirmLeave();
};

async function handleAdminTabClick(event, page, navigate) {
    // Let RouterLink/browser handle middle-click and modifier-key navigation.
    if (event.button !== 0 || event.metaKey || event.altKey || event.ctrlKey || event.shiftKey) {
        return;
    }

    if (route.params.page === page) {
        event.preventDefault();
        return;
    }

    const targetRoute = { name: 'admin', params: { page } };
    if (!(await canLeaveAdminTab(targetRoute, route))) {
        event.preventDefault();
        return;
    }

    goNoTransition();
    navigate(event);
}
</script>

<template>
    <section class="admin-shell container-fluid py-1 py-lg-2">
        <div class="card admin-frame shadow-sm">
            <h5 class="card-header text-center admin-frame-header">{{ $t('admin.header') }}</h5>

            <div class="card-body admin-frame-body">
                <div class="admin-layout">
                    <aside class="admin-sidebar">
                        <div class="card admin-nav-card shadow-sm">
                            <div class="card-body p-3">
                                <nav aria-label="Admin navigation">
                                    <ul class="nav flex-column gap-1">
                                        <li v-for="item in navigationItems" :key="item.page" class="nav-item">
                                            <RouterLink :to="{ name: 'admin', params: { page: item.page } }" custom
                                                v-slot="{ href, navigate, isActive }">
                                                <a :href="href" class="admin-nav-link" :class="{ active: isActive }"
                                                    @click="handleAdminTabClick($event, item.page, navigate)">
                                                    <i :class="item.icon"></i>
                                                    <span>{{ $t(item.labelKey) }}</span>
                                                </a>
                                            </RouterLink>
                                        </li>
                                    </ul>
                                </nav>
                            </div>
                        </div>

                        <div class="card text-center admin-release-card shadow-sm mt-3">
                            <div class="card-body d-flex justify-content-center">
                                <div class="col">
                                    <img src="/logo.png" alt="logo" height="70" class="mb-2"><br>
                                    <small
                                        class="d-inline-flex mb-1 px-2 py-1 fw-semibold text-info-emphasis bg-info-subtle border border-info-subtle rounded-2">
                                        {{ releaseVersion }} ({{ releaseDate }})
                                    </small>
                                    <div class="mb-1 px-2 py-1"></div>
                                    <div class="btn-group mb-1 px-2 py-1">
                                        <a href="https://github.com/Ookami-Git/datalchemist/"
                                            class="btn btn-sm btn-dark border-light-subtle" target="_blank"><i
                                                class="bi bi-github"></i></a>
                                        <a :href="'https://github.com/Ookami-Git/datalchemist/releases/tag/' + releaseVersion"
                                            class="btn btn-sm btn-light border-dark-subtle"
                                            target="_blank">Changelog</a>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </aside>

                    <main class="admin-content">
                        <form class="admin-content-form">
                            <component :is="activeTab" />
                        </form>
                    </main>
                </div>
            </div>
        </div>
    </section>
</template>
