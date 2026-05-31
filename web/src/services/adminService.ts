import apiClient from '@/utils/api'

export interface AdminOverviewStats {
    totalMembers: number
    totalEvents: number
    totalTokens: number
    activeThisMonth: number
    uniqueAttendees: number
    averageAttendance: number
}

export interface AttendanceRecord {
    id: string
    name: string
    submittedBy: string
    participantCount: number
    recorded: boolean
    successful: boolean
    dateCreated: string
}

export interface TokenTransaction {
    id: string
    memberId: string
    amount: number
    reason: string
    createdAt: string
    comment?: string
    giverId?: string
    attendanceId?: string
}

export interface MemberSummary {
    id: string
    username: string
    rank: string
    attendance: number
    tokenBalance: number
    rsiHandle?: string
}

interface PaginatedResponse<T> {
    records?: T[]
    members?: T[]
    page: number
    limit: number
}

class AdminService {
    async getOverview(): Promise<AdminOverviewStats> {
        const response = await apiClient.get<AdminOverviewStats>('/admin/overview')
        return response.data
    }

    async getAttendance(limit: number = 50, page: number = 1): Promise<PaginatedResponse<AttendanceRecord>> {
        const response = await apiClient.get<PaginatedResponse<AttendanceRecord>>('/admin/attendance', {
            params: { limit, page },
        })
        return response.data
    }

    async getTokenLedger(limit: number = 50, page: number = 1): Promise<PaginatedResponse<TokenTransaction>> {
        const response = await apiClient.get<PaginatedResponse<TokenTransaction>>('/admin/token-ledger', {
            params: { limit, page },
        })
        return response.data
    }

    async getMembers(limit: number = 50, page: number = 1, search?: string): Promise<PaginatedResponse<MemberSummary>> {
        const response = await apiClient.get<PaginatedResponse<MemberSummary>>('/admin/members', {
            params: { limit, page, search },
        })
        return response.data
    }
}

export default new AdminService()
