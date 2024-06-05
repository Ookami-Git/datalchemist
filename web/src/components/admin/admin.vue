<script setup>
import { inject } from 'vue';
import global from './global.vue'
import acl from './acl.vue'
import users from './users.vue'
import groups from './groups.vue'
import navbar from './navbar.vue'

const parameters = inject('parameters');

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
                    <br>
                    <div class="card text-center border-0">
                        <div class="card-body d-flex justify-content-center">
                            <div class="col">
                                <img src="/logo.png" alt="logo" height="70" class="mb-2"><br>
                                <small class="d-inline-flex mb-1 px-2 py-1 fw-semibold text-info-emphasis bg-info-subtle border border-info-subtle rounded-2">{{ parameters.release.version }} ({{ parameters.release.date.split('T')[0] }})</small>
                                <div class="mb-1 px-2 py-1"></div>
                                <div class="btn-group mb-1 px-2 py-1">
                                    <a href="https://github.com/Ookami-Git/datalchemist/" class="btn btn-sm btn-dark border-light-subtle" target="_blank"><i class="bi bi-github"></i></a>
                                    <a :href="'https://github.com/Ookami-Git/datalchemist/releases/tag/' + parameters.release.version" class="btn btn-sm btn-light border-dark-subtle" target="_blank">Changelog</a>
                                </div>
                            </div>
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