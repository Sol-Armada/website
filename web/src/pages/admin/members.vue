<script setup lang="ts">
  import { onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type MemberSummary } from '@/services/adminService'
  import { type RealtimeEnvelope, WS_TOPIC_ADMIN_MEMBERS, wsClient } from '@/services/wsClient'

  const loading = ref(true)
  const isRefreshing = ref(false)
  const error = ref<string | null>(null)
  const members = ref<MemberSummary[]>([])
  const search = ref('')
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)

  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let refreshTimer: number | null = null
  let inFlightRequest: Promise<void> | null = null
  let queuedRefreshMode: 'background' | 'blocking' | null = null
  let nextPageLoadMode: 'background' | 'blocking' = 'blocking'
  const unsubscribers: Array<() => void> = []

  function scheduleRefresh() {
    if (search.value.trim() !== '') {
      return
    }

    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
    }
    refreshTimer = window.setTimeout(() => {
      refreshTimer = null
      void loadMembers({ background: true })
    }, 400)
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
    const isBackground = options.background === true

    if (inFlightRequest !== null) {
      queuedRefreshMode = !isBackground || queuedRefreshMode === 'blocking'
        ? 'blocking'
        : 'background'
      await inFlightRequest
      return
    }

    if (isBackground) {
      isRefreshing.value = true
    } else {
      loading.value = true
      error.value = null
    }

    const request = (async() => {
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
    })()

    inFlightRequest = request
    await request
    inFlightRequest = null

    if (queuedRefreshMode !== null) {
      const nextMode = queuedRefreshMode
      queuedRefreshMode = null
      void loadMembers({ background: nextMode === 'background' })
    }
  }

  function goToPreviousPage(): void {
    if (page.value <= 1 || loading.value) return

    page.value -= 1
  }

  function goToFirstPage(): void {
    if (page.value === 1 || loading.value) return

    page.value = 1
  }

  function jumpToPage(): void {
    if (loading.value) return

    const nextPage = Number.parseInt(pageInput.value, 10)
    if (!Number.isFinite(nextPage) || nextPage < 1) {
      pageInput.value = String(page.value)
      return
    }

    if (nextPage === page.value) {
      return
    }

    page.value = nextPage
  }

  function goToNextPage(): void {
    if (!hasNextPage.value || loading.value) return

    page.value += 1
  }

  watch(page, () => {
    pageInput.value = String(page.value)
    const shouldUseBackgroundLoad = nextPageLoadMode === 'background'
    nextPageLoadMode = 'blocking'
    void loadMembers({ background: shouldUseBackgroundLoad })
  })

  watch(search, () => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)

    searchDebounceTimer = setTimeout(() => {
      nextPageLoadMode = 'background'

      if (page.value !== 1) {
        page.value = 1
        return
      }

      void loadMembers({ background: true })
    }, 300)
  })

  onMounted(async() => {
    await loadMembers()
    unsubscribers.push(wsClient.onTopic(WS_TOPIC_ADMIN_MEMBERS, applyRealtimeMemberEvent))
  })

  onBeforeUnmount(() => {
    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
      refreshTimer = null
    }
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  })

  onUnmounted(() => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })
</script>

<template>
  <PortalShell>
    <PageHeader subtitle="Searchable member directory with rank tags and simple paging." title="Members" />

    <DataPanel description="Browse members with search and paging controls." title="Member Directory">
      <input
        v-model="search"
        class="w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface mb-2"
        placeholder="Search members..."
        type="search"
      >

      <br>

      <StatePanel v-if="error" :message="error" title="Members load failed" tone="error" />

      <div v-else-if="members.length > 0" class="overflow-x-auto rounded-lg border border-subtle">
        <table class="w-full text-left text-sm text-on-surface">
          <thead class="bg-surface-variant/40 text-on-surface-variant">
            <tr>
              <th class="px-3 py-2">Username</th>
              <th class="px-3 py-2">Rank</th>
              <th class="px-3 py-2">Attendance</th>
              <th class="px-3 py-2">Tokens</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="member in members" :key="member.id" class="border-t border-subtle">
              <td class="px-3 py-2">{{ member.username }}</td>
              <td class="px-3 py-2">{{ member.rank }}</td>
              <td class="px-3 py-2">{{ member.attendance }}</td>
              <td class="px-3 py-2">{{ member.tokenBalance }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="text-sm text-on-surface-variant">No members found.</p>

      <div class="mt-4 flex items-center justify-between gap-3 text-sm text-on-surface-variant">
        <span>Page {{ page }}</span>

        <div class="flex items-center gap-2">
          <button
            class="rounded-md border border-subtle px-3 py-1.5 transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading || page === 1"
            type="button"
            @click="goToFirstPage"
          >
            First
          </button>

          <button
            class="rounded-md border border-subtle px-3 py-1.5 transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading || page === 1"
            type="button"
            @click="goToPreviousPage"
          >
            Previous
          </button>

          <input
            v-model="pageInput"
            class="w-20 rounded-md border border-subtle bg-transparent px-2 py-1.5 text-right text-sm text-on-surface"
            min="1"
            type="number"
            @keydown.enter.prevent="jumpToPage"
          >

          <button
            class="rounded-md border border-subtle px-3 py-1.5 transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading"
            type="button"
            @click="jumpToPage"
          >
            Go
          </button>

          <button
            class="rounded-md border border-subtle px-3 py-1.5 transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading || !hasNextPage"
            type="button"
            @click="goToNextPage"
          >
            Next
          </button>
        </div>
      </div>
    </DataPanel>
  </PortalShell>
</template>
