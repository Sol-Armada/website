import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface AuthUser {
    id: string
    discordID: string
    username: string
    email: string
    roles: string[]
}

export const useAuthStore = defineStore('auth', () => {
    const user = ref<AuthUser | null>(null)
    const isAuthenticated = computed(() => user.value !== null)

    const setUser = (newUser: AuthUser) => {
        user.value = newUser
    }

    const clearUser = () => {
        user.value = null
    }

    const hasRole = (role: string): boolean => {
        if (!user.value) return false
        return user.value.roles.includes(role)
    }

    return {
        user,
        isAuthenticated,
        setUser,
        clearUser,
        hasRole,
    }
})
