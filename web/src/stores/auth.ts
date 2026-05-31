import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authService, { type AuthUser } from '@/services/authService'

export { type AuthUser } from '@/services/authService'

export const useAuthStore = defineStore('auth', () => {
    const user = ref<AuthUser | null>(null)
    const loading = ref(false)
    const error = ref<string | null>(null)

    const isAuthenticated = computed(() => user.value !== null)

    const setUser = (newUser: AuthUser) => {
        user.value = newUser
        error.value = null
    }

    const clearUser = () => {
        user.value = null
    }

    const hasRole = (role: string): boolean => {
        if (!user.value) return false
        return user.value.roles.includes(role)
    }

    const hasAnyRole = (roles: string[]): boolean => {
        if (!user.value) return false
        return roles.some(role => user.value!.roles.includes(role))
    }

    /**
     * Check if user is authenticated and load user data
     */
    const checkAuth = async (): Promise<boolean> => {
        loading.value = true
        error.value = null

        try {
            const userData = await authService.checkAuth()
            if (userData) {
                setUser(userData)
                return true
            } else {
                clearUser()
                return false
            }
        } catch (err: any) {
            error.value = err.message || 'Failed to check authentication'
            clearUser()
            return false
        } finally {
            loading.value = false
        }
    }

    /**
     * Initiate Discord OAuth login flow
     */
    const login = async (): Promise<void> => {
        loading.value = true
        error.value = null

        try {
            const loginUrl = await authService.getLoginUrl()
            // Redirect to Discord OAuth
            window.location.href = loginUrl
        } catch (err: any) {
            error.value = err.message || 'Failed to initiate login'
            loading.value = false
        }
    }

    /**
     * Logout the current user
     */
    const logout = async (): Promise<void> => {
        loading.value = true
        error.value = null

        try {
            await authService.logout()
            clearUser()
        } catch (err: any) {
            error.value = err.message || 'Failed to logout'
        } finally {
            loading.value = false
        }
    }

    /**
     * Fetch current user info
     */
    const fetchUser = async (): Promise<void> => {
        loading.value = true
        error.value = null

        try {
            const userData = await authService.me()
            setUser(userData)
        } catch (err: any) {
            error.value = err.message || 'Failed to fetch user'
            clearUser()
            throw err
        } finally {
            loading.value = false
        }
    }

    return {
        user,
        loading,
        error,
        isAuthenticated,
        setUser,
        clearUser,
        hasRole,
        hasAnyRole,
        checkAuth,
        login,
        logout,
        fetchUser,
    }
})
