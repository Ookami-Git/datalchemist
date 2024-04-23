<script setup>
import global from './global.vue'
import acl from './acl.vue'
import users from './users.vue'
import groups from './groups.vue'
import navbar from './navbar.vue'

const tabs = {
    global,
    acl,
    users,
    groups,
    navbar
}
</script>

<template>
<div class="row">
    <div class="col-md-12">
      <div class="card">
        <h5 class="card-header text-center">{{ $t('admin.header') }}</h5>
        <div class="card-body">
            <div class="row">
                <div class="col-2">
                    <div class="card">
                        <div class="card-body">
                            <nav class="navbar">
                                <ul class="navbar-nav flex-column">
                                <li class="nav-item">
                                    <RouterLink class="nav-link" :to="{ name: 'admin', params: { page: 'global' } }" active-class="active"><i class="bi bi-card-checklist"></i> {{ $t('admin.global.header') }}</RouterLink>
                                </li>
                                <li class="nav-item">
                                    <RouterLink class="nav-link" :to="{ name: 'admin', params: { page: 'navbar' } }" active-class="active"><i class="bi bi-window"></i> {{ $t('admin.navbar.header') }}</RouterLink>
                                </li>
                                <li class="nav-item">
                                    <RouterLink class="nav-link" :to="{ name: 'admin', params: { page: 'users' } }" active-class="active"><i class="bi bi-people-fill"></i> {{ $t('admin.users.header') }}</RouterLink>
                                </li>
                                <li class="nav-item">
                                    <RouterLink class="nav-link" :to="{ name: 'admin', params: { page: 'groups' } }" active-class="active"><i class="bi bi-collection-fill"></i> {{ $t('admin.groups.header') }}</RouterLink>
                                </li>
                                <li class="nav-item">
                                    <RouterLink class="nav-link" :to="{ name: 'admin', params: { page: 'acl' } }" active-class="active"><i class="bi bi-shield-lock-fill"></i> {{ $t('admin.acl.header') }}</RouterLink>
                                </li>
                                </ul>
                            </nav>
                        </div>
                    </div>
                </div>
                <div class="col-10">
                    <form>
                        <component :is="tabs[$route.params.page]" />
                    </form>
                </div>
            </div>
        </div>
      </div>
    </div>
</div>
</template>