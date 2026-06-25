import type { TokenTransaction } from '@/services/adminService'
import { requestJson } from '@/services/http'

export interface MemberActivity {
  type: string
  title: string
  date: string
}

interface PaginatedResponse<T> {
  records?: T[]
  page: number
  limit: number
}

export interface MemberDashboardData {
  attendance: number
  tokens: number
  rank: string
  tokenLedger: TokenTransaction[]
  recentActivity: MemberActivity[]
}

export interface MemberProfileData {
  id: string
  discordID: string
  username: string
  email: string
  roles: string[]
  rank: string
  attendanceCount: number
  tokensBalance: number
  memberSince?: string
  rsiHandle?: string
  validated: boolean
}

export const memberService = {
  async getDashboard(): Promise<MemberDashboardData> {
    return requestJson<MemberDashboardData>('/api/member/dashboard')
  },

  async getTokenLedger(limit = 50, page = 1): Promise<PaginatedResponse<TokenTransaction>> {
    return requestJson<PaginatedResponse<TokenTransaction>>('/api/member/token-ledger', undefined, {
      limit,
      page,
    })
  },

  async getProfile(): Promise<MemberProfileData> {
    return requestJson<MemberProfileData>('/api/member/profile')
  },
}
