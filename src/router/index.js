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
    const appStore = useAppStore()

    // redirect to login page if not logged in and trying to access a restricted page
    if (to.path !== '/login' && !localStorage.getItem('logged_in')) {
        next('/login')
        return
    }

    // redirect to onboarding page if logged in and not onboarded
    if (to.path !== '/onboard' && localStorage.getItem('logged_in') && !localStorage.getItem('onboarded')) {
        next('/onboard')
        return
    }

    // if (to.path === '/onboard' && localStorage.getItem('logged_in') && localStorage.getItem('onboarded')) {
    //     next('/')
    //     return
    // }

    // check if they have permission to the page
    if (to.meta.requiresOfficer && appStore.me.rank.id > 3) {
        next('/')
        return
    }

    next()
})

// router.afterEach((to, from) => {
//     const appStore = useAppStore()
//     const code = new URLSearchParams(window.location.search).get('code')

//     if (to.path === '/login' && code) {
//         appStore.login(code).then((res) => {
//             if (res) {
//                 window.location.href = "/"
//             }
//         }).catch((err) => {
//             if (err === 'invalid_grant') {
//                 // clear the code from the url
//                 window.history.replaceState({}, document.title, "/login")
//             }
//         })
//     }
// })

export default router
