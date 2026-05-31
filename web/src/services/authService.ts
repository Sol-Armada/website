import apiClient from '@/utils/api'

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
     * Get the Discord OAuth login URL
     */
    async getLoginUrl(): Promise<string> {
        try {
            const response = await apiClient.get('/auth/login', {
                maxRedirects: 0,
                validateStatus: (status) => status === 307,
            })
            return response.headers.location
        } catch (error: any) {
            if (error.response?.status === 307) {
                return error.response.headers.location
            }
            throw new Error('Failed to get login URL')
        }
    }

    /**
     * Get current user info (if authenticated)
     */
    async me(): Promise<AuthUser> {
        const response = await apiClient.get<AuthUser>('/auth/me')
        return response.data
    }

    /**
     * Logout the current user
     */
    async logout(): Promise<void> {
        await apiClient.post('/auth/logout')
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
