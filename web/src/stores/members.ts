import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminService, type MemberSummary } from '@/services/adminService'
import { type RealtimeEnvelope, WS_TOPIC_ADMIN_MEMBERS, wsClient } from '@/services/wsClient'
import { createRefreshScheduler } from '@/stores/refreshScheduler'
import { createRequestQueue } from '@/stores/requestQueue'

export const useMembersStore = defineStore('members', () => {
  const loading = ref(true)
  const isRefreshing = ref(false)
  const error = ref<string | null>(null)
  const members = ref<MemberSummary[]>([])
  const search = ref('')
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)

  const refreshScheduler = createRefreshScheduler()
  const membersRequestQueue = createRequestQueue()
  let unsubscribeMembers: (() => void) | null = null

  function scheduleRefresh() {
    if (search.value.trim() !== '') {
      return
    }

    refreshScheduler.schedule(() => {
      void loadMembers({ background: true })
    })
  }

  function logRealtimeMemberDecision(decision: string, event: RealtimeEnvelope): void {
    if (!import.meta.env.DEV) {
      return
    }

    const operation = String(event.payload?.operation ?? '').toLowerCase()
    const memberID = String(event.payload?.member_id ?? event.payload?.primary_key ?? '').trim()
    console.debug('[members:realtime]', {
      decision,
      operation,
      memberID,
      page: page.value,
      pageSize: limit.value,
      visibleRows: members.value.length,
      searchActive: search.value.trim() !== '',
      sequence: event.sequence,
    })
  }

  function applyRealtimeMemberEvent(event: RealtimeEnvelope): void {
    if (search.value.trim() !== '') {
      logRealtimeMemberDecision('ignored-search-active', event)
      return
    }

    const operation = String(event.payload?.operation ?? '').toLowerCase()
    const memberID = String(event.payload?.member_id ?? event.payload?.primary_key ?? '').trim()
    const payloadMember = event.payload?.member as MemberSummary | undefined

    if (operation === 'delete' && memberID !== '') {
      const nextMembers = members.value.filter(member => member.id !== memberID)
      if (nextMembers.length !== members.value.length) {
        members.value = nextMembers
        logRealtimeMemberDecision('patched-delete', event)
        return
      }
      logRealtimeMemberDecision('ignored-delete-not-visible', event)
      return
    }

    if (!payloadMember || !payloadMember.id) {
      logRealtimeMemberDecision('fallback-refresh-missing-member-payload', event)
      scheduleRefresh()
      return
    }

    const existingIndex = members.value.findIndex(member => member.id === payloadMember.id)
    if (existingIndex !== -1) {
      const nextMembers = [...members.value]
      nextMembers[existingIndex] = payloadMember
      members.value = nextMembers
      logRealtimeMemberDecision('patched-update-visible-row', event)
      return
    }

    if (operation === 'insert' && page.value === 1 && members.value.length < limit.value) {
      members.value = [payloadMember, ...members.value]
      logRealtimeMemberDecision('patched-insert-page-1', event)
      return
    }

    logRealtimeMemberDecision('fallback-refresh-not-visible', event)
    scheduleRefresh()
  }

  async function loadMembers(options: { background?: boolean } = {}): Promise<void> {
    await membersRequestQueue.run(options, async isBackground => {
      if (isBackground) {
        isRefreshing.value = true
      } else {
        loading.value = true
        error.value = null
      }

      try {
        const response = await adminService.getMembers(limit.value, page.value, search.value || undefined)
        members.value = response.members || []
        hasNextPage.value = members.value.length === limit.value
        error.value = null
      } catch(error_: any) {
        if (!isBackground || members.value.length === 0) {
          error.value = error_?.message || 'Failed to load members'
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

  async function initialize(): Promise<void> {
    if (unsubscribeMembers === null) {
      unsubscribeMembers = wsClient.onTopic(WS_TOPIC_ADMIN_MEMBERS, applyRealtimeMemberEvent)
    }

    await loadMembers()
  }

  function dispose(): void {
    refreshScheduler.clear()

    membersRequestQueue.clear()

    if (unsubscribeMembers !== null) {
      unsubscribeMembers()
      unsubscribeMembers = null
    }
  }

  return {
    loading,
    isRefreshing,
    error,
    members,
    search,
    page,
    pageInput,
    hasNextPage,
    loadMembers,
    initialize,
    dispose,
  }
})
