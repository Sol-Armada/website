import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminService, type AttendanceRecord, type MemberSummary } from '@/services/adminService'
import { WS_TOPIC_ADMIN_ATTENDANCE, wsClient } from '@/services/wsClient'
import { createRefreshScheduler } from '@/stores/refreshScheduler'
import { createRequestQueue } from '@/stores/requestQueue'

export const useAttendanceStore = defineStore('attendance', () => {
  const loading = ref(true)
  const isRefreshing = ref(false)
  const error = ref<string | null>(null)
  const records = ref<AttendanceRecord[]>([])
  const search = ref('')
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)

  const availableAttendanceNames = ref<string[]>([])
  const availableMembers = ref<Record<string, string>>({})
  const memberSearchLoading = ref(false)
  const memberSearchResults = ref<MemberSummary[]>([])
  const managerSearchLoading = ref(false)
  const managerSearchResults = ref<MemberSummary[]>([])

  const refreshScheduler = createRefreshScheduler()
  const attendanceRequestQueue = createRequestQueue()
  let unsubscribeAttendance: (() => void) | null = null

  async function loadAttendance(options: { background?: boolean } = {}): Promise<void> {
    await attendanceRequestQueue.run(options, async isBackground => {
      if (isBackground) {
        isRefreshing.value = true
      } else {
        loading.value = true
        error.value = null
      }

      try {
        const response = await adminService.getAttendance(limit.value, page.value, search.value || undefined)
        records.value = response.records || []
        hasNextPage.value = records.value.length === limit.value
        error.value = null
      } catch(error_: any) {
        if (!isBackground || records.value.length === 0) {
          error.value = error_?.message || 'Failed to load attendance records'
          hasNextPage.value = false
        }
      } finally {
        if (isBackground) {
          isRefreshing.value = false
        } else {
          loading.value = false
        }
      }
    })
  }

  function scheduleRefresh(): void {
    if (search.value.trim() !== '') {
      return
    }

    refreshScheduler.schedule(() => {
      void loadAttendance({ background: true })
    })
  }

  async function initialize(): Promise<void> {
    if (unsubscribeAttendance === null) {
      unsubscribeAttendance = wsClient.onTopic(WS_TOPIC_ADMIN_ATTENDANCE, scheduleRefresh)
    }

    await loadAttendance()
  }

  function dispose(): void {
    refreshScheduler.clear()

    attendanceRequestQueue.clear()

    if (unsubscribeAttendance !== null) {
      unsubscribeAttendance()
      unsubscribeAttendance = null
    }
  }

  async function loadCreateModalOptions(): Promise<void> {
    try {
      const [names, membersResponse] = await Promise.all([
        adminService.getAvailableAttendanceNames(),
        adminService.getMembers(100, 1),
      ])

      availableAttendanceNames.value = names
      memberSearchResults.value = membersResponse.members || []
      managerSearchResults.value = membersResponse.members || []
      availableMembers.value = memberSearchResults.value.reduce<Record<string, string>>((acc, member) => {
        acc[member.id] = member.username
        return acc
      }, {})
    } catch(error_) {
      availableAttendanceNames.value = []
      memberSearchResults.value = []
      managerSearchResults.value = []
      console.error('Failed to fetch modal options:', error_)
    }
  }

  async function searchMembers(query: string): Promise<void> {
    if (query === '') {
      return
    }

    memberSearchLoading.value = true
    try {
      const response = await adminService.getMembers(100, 1, query || undefined)
      const members = response.members || []
      memberSearchResults.value = members

      const nextMap = { ...availableMembers.value }
      for (const member of members) {
        nextMap[member.id] = member.username
      }
      availableMembers.value = nextMap
    } catch(error_) {
      memberSearchResults.value = []
      console.error('Failed to search members:', error_)
    } finally {
      memberSearchLoading.value = false
    }
  }

  async function searchManagers(query: string): Promise<void> {
    if (query === '') {
      return
    }

    managerSearchLoading.value = true
    try {
      const response = await adminService.getMembers(100, 1, query || undefined)
      const members = response.members || []
      managerSearchResults.value = members

      const nextMap = { ...availableMembers.value }
      for (const member of members) {
        nextMap[member.id] = member.username
      }
      availableMembers.value = nextMap
    } catch(error_) {
      managerSearchResults.value = []
      console.error('Failed to search managers:', error_)
    } finally {
      managerSearchLoading.value = false
    }
  }

  return {
    loading,
    isRefreshing,
    error,
    records,
    search,
    page,
    pageInput,
    limit,
    hasNextPage,
    availableAttendanceNames,
    availableMembers,
    memberSearchLoading,
    memberSearchResults,
    managerSearchLoading,
    managerSearchResults,
    loadAttendance,
    scheduleRefresh,
    initialize,
    dispose,
    loadCreateModalOptions,
    searchMembers,
    searchManagers,
  }
})
