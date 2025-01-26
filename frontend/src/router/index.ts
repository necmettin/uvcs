import { createRouter, createWebHistory } from 'vue-router';
import { api } from '../services/api';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'home',
            component: () => import('../views/HomeView.vue'),
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('../views/LoginView.vue'),
            meta: { requiresGuest: true },
        },
        {
            path: '/register',
            name: 'register',
            component: () => import('../views/RegisterView.vue'),
            meta: { requiresGuest: true },
        },
        {
            path: '/repository/:name',
            name: 'repository',
            component: () => import('../views/RepositoryView.vue'),
            meta: { requiresAuth: true },
            props: true,
        },
        {
            path: '/repository/:name/commits',
            name: 'commits',
            component: () => import('../views/CommitsView.vue'),
            meta: { requiresAuth: true },
            props: true,
        },
        {
            path: '/repository/:name/branches',
            name: 'branches',
            component: () => import('../views/BranchesView.vue'),
            meta: { requiresAuth: true },
            props: true,
        },
        {
            path: '/repository/:name/settings',
            name: 'repository-settings',
            component: () => import('../views/RepositorySettingsView.vue'),
            meta: { requiresAuth: true },
            props: true,
        },
    ],
});

// Navigation guards
router.beforeEach((to, _, next) => {
    const isAuthenticated = api.isAuthenticated();

    if (to.meta.requiresAuth && !isAuthenticated) {
        next({ name: 'login', query: { redirect: to.fullPath } });
    } else if (to.meta.requiresGuest && isAuthenticated) {
        next({ name: 'home' });
    } else {
        next();
    }
});

export default router; 