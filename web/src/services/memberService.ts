import { requestJson } from '@/services/http'

export interface MemberActivity {
  type: string
  title: string
  date: string
}

export interface MemberDashboardData {
  attendance: number
  tokens: number
  rank: string
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
}

export const memberService = {
  async getDashboard(): Promise<MemberDashboardData> {
    return requestJson<MemberDashboardData>('/api/member/dashboard')
  },

  async getProfile(): Promise<MemberProfileData> {
    return requestJson<MemberProfileData>('/api/member/profile')
  },
}
