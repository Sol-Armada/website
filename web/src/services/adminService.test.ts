import { describe, expect, it, vi, beforeEach } from 'vitest'

const { mockGet } = vi.hoisted(() => ({
    mockGet: vi.fn(),
}))

vi.mock('@/utils/api', () => ({
    default: {
        get: mockGet,
    },
}))

import adminService from './adminService'

describe('adminService', () => {
    beforeEach(() => {
        mockGet.mockReset()
    })

    it('fetches overview stats', async () => {
        const payload = {
            totalMembers: 10,
            totalEvents: 5,
            totalTokens: 120,
            activeThisMonth: 8,
            uniqueAttendees: 7,
            averageAttendance: 3,
        }
        mockGet.mockResolvedValueOnce({ data: payload })

        const result = await adminService.getOverview()

        expect(mockGet).toHaveBeenCalledWith('/admin/overview')
        expect(result).toEqual(payload)
    })

    it('passes pagination params for attendance', async () => {
        const payload = { records: [], page: 2, limit: 25 }
        mockGet.mockResolvedValueOnce({ data: payload })

        const result = await adminService.getAttendance(25, 2)

        expect(mockGet).toHaveBeenCalledWith('/admin/attendance', {
            params: { limit: 25, page: 2 },
        })
        expect(result).toEqual(payload)
    })

    it('passes search params for members', async () => {
        const payload = { members: [], page: 1, limit: 50 }
        mockGet.mockResolvedValueOnce({ data: payload })

        const result = await adminService.getMembers(50, 1, 'doug')

        expect(mockGet).toHaveBeenCalledWith('/admin/members', {
            params: { limit: 50, page: 1, search: 'doug' },
        })
        expect(result).toEqual(payload)
    })
})
