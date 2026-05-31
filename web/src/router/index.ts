import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Placeholder pages - to be created in Phase 4-6
const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/auth/login',
    },
    {
        path: '/auth/login',
        name: 'Login',
        component: () => import('@/pages/auth/LoginPage.vue'),
        meta: { requiresAuth: false },
    },
    {
        path: '/auth/callback',
        name: 'Callback',
        component: () => import('@/pages/auth/CallbackPage.vue'),
        meta: { requiresAuth: false },
    },
    {
        path: '/dashboard',
        component: () => import('@/layouts/AppLayout.vue'),
        children: [
            {
                path: '',
                name: 'MemberDashboard',
                component: () => import('@/pages/member/DashboardPage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['member'] },
            },
            {
                path: 'profile',
                name: 'MemberProfile',
                component: () => import('@/pages/member/ProfilePage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['member'] },
            },
        ],
    },
    {
        path: '/admin',
        component: () => import('@/layouts/AppLayout.vue'),
        children: [
            {
                path: 'overview',
                name: 'AdminOverview',
                component: () => import('@/pages/admin/OverviewPage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['admin'] },
            },
            {
                path: 'attendance',
                name: 'AdminAttendance',
                component: () => import('@/pages/admin/AttendancePage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['moderator', 'admin'] },
            },
            {
                path: 'token-ledger',
                name: 'AdminTokenLedger',
                component: () => import('@/pages/admin/TokenLedgerPage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['admin'] },
            },
            {
                path: 'members',
                name: 'AdminMembers',
                component: () => import('@/pages/admin/MembersPage.vue'),
                meta: { requiresAuth: true, requiredRoles: ['admin'] },
            },
        ],
    },
]

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
})

router.beforeEach((to, _from, next) => {
    const authStore = useAuthStore()
    const requiresAuth = to.meta.requiresAuth as boolean
    const requiredRoles = (to.meta.requiredRoles as string[]) || []

    if (requiresAuth) {
        if (!authStore.isAuthenticated) {
            next('/auth/login')
            return
        }

        if (requiredRoles.length > 0) {
            const hasRole = requiredRoles.some((role) => authStore.hasRole(role))
            if (!hasRole) {
                next('/dashboard')
                return
            }
        }
    }

    next()
})

export default router
