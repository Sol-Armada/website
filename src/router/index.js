/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router/auto'
import { setupLayouts } from 'virtual:generated-layouts'
import { useAppStore } from '@/stores/app'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    extendRoutes: setupLayouts,
})

router.beforeEach((to, from, next) => {
    // redirect to login page if not logged in and trying to access a restricted page
    if (to.path !== '/login' && !localStorage.getItem('logged_in')) {
        next('/login')
        return
    }

    // check if they have permission to the page
    const appStore = useAppStore()
    if (to.meta.requiresOfficer && appStore.me.rank.id > 3) {
        next('/')
        return
    }

    next()
})

router.afterEach((to, from) => {
    const appStore = useAppStore()
    const code = new URLSearchParams(window.location.search).get('code')

    if (to.path === '/login') {
        appStore.login(code)
    }
})

export default router
