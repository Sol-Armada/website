/**
 * router/index.ts
 *
 * Automatic routes for ./src/pages/*.vue
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router'
import { routes } from 'vue-router/auto-routes'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async to => {
  const authStore = useAuthStore()

  if (to.path === '/') {
    return '/dashboard'
  }

  if (to.path.startsWith('/auth')) {
    if (authStore.isAuthenticated && to.path === '/auth/login') {
      return '/dashboard'
    }
    return true
  }

  const isProtected = to.path.startsWith('/dashboard') || to.path.startsWith('/admin') || to.path.startsWith('/projects')

  if (isProtected && !authStore.isAuthenticated) {
    const isAuthenticated = await authStore.checkAuth()
    if (!isAuthenticated) {
      return '/auth/login'
    }
  }

  // Projects page is admin-only
  if (to.path.startsWith('/projects') && !authStore.hasRole('admin')) {
    return '/dashboard'
  }

  if (to.path.startsWith('/admin/attendance') && !authStore.hasAnyRole(['moderator', 'admin'])) {
    return '/dashboard'
  }

  if (to.path.startsWith('/admin') && !to.path.startsWith('/admin/attendance') && !authStore.hasRole('admin')) {
    return '/dashboard'
  }

  return true
})

export default router
