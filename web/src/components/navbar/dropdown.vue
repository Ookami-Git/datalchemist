<script setup>
import dropitem from './dropitem.vue'
const props = defineProps(['menuitem']);
</script>

<template>
    <li class='nav-item dropdown'>
        <a class='nav-link dropdown-toggle' href='#' role='button' data-bs-toggle='dropdown' aria-expanded='false'>{{ menuitem.name }}</a>
        <ul class='dropdown-menu'>
            <dropitem v-for="(subitem, index) in menuitem.subitems" :key="index" :menuitem="subitem"></dropitem>
        </ul>
    </li>
</template>