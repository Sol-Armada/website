import { requestJson } from '@/services/http'

export interface AdminOverviewData {
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
  awardTokens: boolean
}

export interface CreateAttendanceRecordRequest {
  submittedBy: string | null
  name: string
  participantIds: string[]
  managerIds: string[]
  awardTokens?: boolean
}

export interface TokenTransaction {
  id: string
  memberId: string
  memberName: string
  amount: number
  reason: string
  createdAt: string
  comment?: string
  giverId?: string
  attendanceId?: string
  attendanceName?: string
}

export interface TokenPeriodAnalytics {
  windowStart: string
  windowEnd: string
  totalEarnings: number
  totalSpending: number
  netAmount: number
  averageEarningPerMember: number
  averageSpendingPerMember: number
  averageEarningPerTransaction: number
  averageSpendingPerTransaction: number
  earningTransactionCount: number
  spendingTransactionCount: number
  earningMemberCount: number
  spendingMemberCount: number
}

export interface TokenReasonAggregation {
  reason: string
  transactionCount: number
  netAmount: number
  totalEarnings: number
  totalSpending: number
}

export interface TokenLedgerAnalytics {
  week: TokenPeriodAnalytics
  month: TokenPeriodAnalytics
  reasons: TokenReasonAggregation[]
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

export const adminService = {
  async getOverview(): Promise<AdminOverviewData> {
    return requestJson<AdminOverviewData>('/api/admin/overview')
  },

  async getAttendance(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<AttendanceRecord>> {
    return requestJson<PaginatedResponse<AttendanceRecord>>('/api/admin/attendance', undefined, {
      limit,
      page,
      search,
    })
  },

  async getAvailableAttendanceNames(): Promise<string[]> {
    return requestJson<string[]>('/api/admin/attendance-names')
  },

  async createAttendanceRecord(payload: CreateAttendanceRecordRequest): Promise<AttendanceRecord> {
    return requestJson<AttendanceRecord>('/api/admin/attendance', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async getTokenLedger(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<TokenTransaction>> {
    return requestJson<PaginatedResponse<TokenTransaction>>('/api/admin/token-ledger', undefined, {
      limit,
      page,
      search,
    })
  },

  async getTokenLedgerAnalytics(): Promise<TokenLedgerAnalytics> {
    return requestJson<TokenLedgerAnalytics>('/api/admin/token-ledger/analytics')
  },

  async getMembers(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<MemberSummary>> {
    return requestJson<PaginatedResponse<MemberSummary>>('/api/admin/members', undefined, {
      limit,
      page,
      search,
    })
  },
}
