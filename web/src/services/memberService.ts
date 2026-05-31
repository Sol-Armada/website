import apiClient from '@/utils/api'

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

class MemberService {
    async getDashboard(): Promise<MemberDashboardData> {
        const response = await apiClient.get<MemberDashboardData>('/member/dashboard')
        return response.data
    }

    async getProfile(): Promise<MemberProfileData> {
        const response = await apiClient.get<MemberProfileData>('/member/profile')
        return response.data
    }
}

export default new MemberService()
