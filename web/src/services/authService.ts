import type { AuthUser } from '@/stores/auth'
import { requestJson, requestNoContent } from '@/services/http'

interface BackendAuthUser {
  id: string
  discordID: string
  username: string
  email: string
  avatar?: string
  roles: string[]
}

function toDisplayRank(roles: string[]): string {
  if (roles.includes('admin')) {
    return 'Admiral'
  }

  if (roles.includes('moderator')) {
    return 'Lieutenant'
  }

  return 'Member'
}

function normalizeUser(user: BackendAuthUser): AuthUser {
  return {
    id: user.id,
    discordID: user.discordID,
    username: user.username,
    email: user.email,
    avatar: user.avatar,
    roles: user.roles,
    displayRank: toDisplayRank(user.roles),
  }
}

export const authService = {
  login(): void {
    window.location.assign('/auth/login')
  },

  async me(): Promise<AuthUser> {
    const user = await requestJson<BackendAuthUser>('/auth/me')
    return normalizeUser(user)
  },

  async logout(): Promise<void> {
    await requestNoContent('/auth/logout', {
      method: 'POST',
    })
  },

  async checkAuth(): Promise<AuthUser | null> {
    try {
      return await this.me()
    } catch {
      return null
    }
  },
}
