import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/authStore'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/login',
            name: 'Login',
            component: () => import('../views/Login.vue'),
            meta: { guest: true }
        },
        {
            path: '/register',
            name: 'Register',
            component: () => import('../views/Register.vue'),
            meta: { guest: true }
        },
        {
            path: '/',
            name: 'Home',
            component: () => import('../views/Dashboard.vue'),
            meta: { auth: true }
        },
        {
            path: '/shared/:token',
            name: 'SharedItem',
            component: () => import('../views/SharedVaultItem.vue'),
            meta: { auth: true }
        }
    ]
})

router.beforeEach((to, _, next) => {
    const { isAuthenticated } = useAuthStore()

    if (to.meta.auth && !isAuthenticated.value) {
        next('/login')
    } else if (to.meta.guest && isAuthenticated.value) {
        next('/')
    } else {
        next()
    }
})

export default router
