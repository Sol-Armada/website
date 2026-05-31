import apiClient from '@/utils/api'
import axios from 'axios'

const authClient = axios.create({
    baseURL: '',
    withCredentials: true,
})

export interface AuthUser {
    id: string
    discordID: string
    username: string
    email: string
    avatar?: string
    roles: string[]
}

export interface AuthResponse {
    user: AuthUser
    csrf: string
}

export interface ErrorResponse {
    error: string
    message: string
}

class AuthService {
    /**
     * Start the Discord OAuth login flow via full-page redirect.
     * Browser navigation handles backend and Discord redirects more reliably than XHR.
     */
    login(): void {
        window.location.assign('/auth/login')
    }

    /**
     * Get current user info (if authenticated)
     */
    async me(): Promise<AuthUser> {
        const response = await authClient.get<AuthUser>('/auth/me')
        return response.data
    }

    /**
     * Logout the current user
     */
    async logout(): Promise<void> {
        await authClient.post('/auth/logout')
    }

    /**
     * Check if user is authenticated by fetching current user
     */
    async checkAuth(): Promise<AuthUser | null> {
        try {
            return await this.me()
        } catch (error) {
            return null
        }
    }
}

export default new AuthService()
