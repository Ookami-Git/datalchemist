import { createRouter, createWebHashHistory } from 'vue-router'
import { i18n } from '@/main.js'

import login from '@/components/login.vue'
import admin from '@/components/admin/admin.vue'
import edit from '@/components/edit/edit.vue'
import editsource from '@/components/edit/source.vue'
import edititem from '@/components/edit/item.vue'
import editview from '@/components/edit/view.vue'
import profil from '@/components/profil.vue'
import view from '@/components/view.vue'
import item from '@/components/view/preview.vue'
import home from '@/components/home.vue'
import notfound from '@/components/error/error404.vue'
import unauthorized from '@/components/error/error401.vue'

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            name: "login",
            path: '/login', 
            component: login
        },
        {
            name: "home",
            path: '/', 
            component: home,
            props: true,
            meta: { requiresAuth: true }
        },
        {
            name: "profil",
            path: '/profile', 
            component: profil,
            meta: { requiresAuth: true }
        },
        {
            name: "admin",
            path: '/admin/:page', 
            component: admin,
            props: true,
            meta: { requiresAuth: true, requiresAdmin: true }
        },
        {
            name: "edit",
            path: '/edit', 
            component: edit,
            meta: { requiresAuth: true, requiresAdmin: true }
        },
        {
            name: "editsource",
            path: '/edit/source/:sourceid', 
            component: editsource,
            props: true,
            meta: { requiresAuth: true, requiresAdmin: true }
        },
        {
            name: "edititem",
            path: '/edit/item/:itemid', 
            component: edititem,
            props: true,
            meta: { requiresAuth: true, requiresAdmin: true }
        },
        {
            name: "editview",
            path: '/edit/view/:viewid', 
            component: editview,
            props: true,
            meta: { requiresAuth: true, requiresAdmin: true }
        },
        {
            name: "view",
            path: '/view/:viewid', 
            component: view,
            props: true,
            meta: { requiresAuth: true }
        },
        {
            name: "item",
            path: '/item/:itemid', 
            component: item,
            props: true,
            meta: { requiresAuth: true }
        },
        {
            name: "error404",
            path: '/:pathMatch(.*)', 
            component: notfound
        },
        {
            name: "unauthorized",
            path: '/access-denied', 
            component: unauthorized
        }
    ]
})

router.beforeEach(async(to) => {
    i18n.locale = localStorage.getItem('language') || 'en'

    // Request the authentication status from the server
    const auth = await fetch(`${window.location.origin}${window.location.pathname}api/auth/status`);
    if (auth.ok) {
        if (to.meta.requiresAdmin) {
            const responseAdmin = await fetch(`${window.location.origin}${window.location.pathname}api/auth/isadmin`);
            const json = await responseAdmin.json();
            if (!json.admin) {
                return { name: 'unauthorized' };
            }
        }
        if (to.path === '/login' || to.path === '/redirect') {
            const redirectPath = localStorage.getItem('redirectPath') || '/';
            localStorage.removeItem('redirectPath'); // Supprimez le chemin de redirection après utilisation
            localStorage.setItem('reloadparameters', true);
            return redirectPath;
        }
    } else {
        if (to.meta.requiresAuth) {
            localStorage.setItem('reloadparameters', true);
            localStorage.setItem('redirectPath', to.fullPath);
            return '/login';
        }
        if (to.name === 'unauthorized') {
            return '/';
        }
    }
    return true;
})

export default router;