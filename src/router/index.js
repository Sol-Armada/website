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
        console.log("NOT LOGGED IN")
        next('/login')
        return
    }

    // redirect to home page if logged in and trying to access login page
    // if (to.path === '/login' && localStorage.getItem('logged_in')) {
    //     next('/')
    //     return
    // }

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
