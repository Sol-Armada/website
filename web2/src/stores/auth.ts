import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authService } from '@/services/authService'
import { wsClient } from '@/services/wsClient'

export type Role = string

export interface AuthUser {
  id: string
  discordID?: string
  username: string
  email?: string
  avatar?: string
  displayRank: string
  roles: Role[]
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => user.value !== null)

  const hasRole = (role: Role): boolean => {
    return user.value?.roles.includes(role) ?? false
  }

  const hasAnyRole = (roles: Role[]): boolean => {
    return roles.some(role => hasRole(role))
  }

  const checkAuth = async(): Promise<boolean> => {
    loading.value = true
    error.value = null

    try {
      const sessionUser = await authService.checkAuth()
      user.value = sessionUser
      return sessionUser !== null
    } catch(error_: any) {
      user.value = null
      error.value = error_?.message || 'Failed to check authentication'
      return false
    } finally {
      loading.value = false
    }
  }

  const login = async(): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      authService.login()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to start login flow'
      loading.value = false
    }
  }

  const logout = async(): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await authService.logout()
      wsClient.disconnect()
      user.value = null
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to log out'
    } finally {
      loading.value = false
    }
  }

  const fetchUser = async(): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      user.value = await authService.me()
    } catch(error_: any) {
      user.value = null
      error.value = error_?.message || 'Failed to fetch user'
      throw error_
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    checkAuth,
    hasRole,
    hasAnyRole,
    login,
    logout,
    fetchUser,
  }
})
