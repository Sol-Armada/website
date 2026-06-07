<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type AttendanceRecord } from '@/services/adminService'
  import { WS_TOPIC_ADMIN_ATTENDANCE, wsClient } from '@/services/wsClient'

  const loading = ref(true)
  const isRefreshing = ref(false)
  const error = ref<string | null>(null)
  const records = ref<AttendanceRecord[]>([])
  const search = ref('')
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let refreshTimer: number | null = null
  let inFlightRequest: Promise<void> | null = null
  let queuedRefreshMode: 'background' | 'blocking' | null = null
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
      void loadAttendance({ background: true })
    }, 400)
  }

  async function loadAttendance(options: { background?: boolean } = {}): Promise<void> {
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
    })()

    inFlightRequest = request
    await request
    inFlightRequest = null

    if (queuedRefreshMode !== null) {
      const nextMode = queuedRefreshMode
      queuedRefreshMode = null
      void loadAttendance({ background: nextMode === 'background' })
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
    void loadAttendance()
  })

  watch(search, () => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }

    searchDebounceTimer = setTimeout(() => {
      page.value = 1
      void loadAttendance({ background: true })
    }, 300)
  })

  onMounted(async() => {
    await loadAttendance()
    unsubscribers.push(wsClient.onTopic(WS_TOPIC_ADMIN_ATTENDANCE, scheduleRefresh))
  })

  onBeforeUnmount(() => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }
    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
      refreshTimer = null
    }
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  })
</script>

<template>
  <PortalShell>
    <PageHeader subtitle="Attendance records list with simple paging controls." title="Attendance" />

    <DataPanel description="Review attendance records and page through history." title="Attendance Records">
      <input
        v-model="search"
        class="w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
        placeholder="Search attendance..."
        type="search"
      >

      <div class="mb-3 mt-2 h-0.5 w-full overflow-hidden rounded-full bg-surface-variant/40">
        <div
          class="h-full w-full bg-primary/80 transition-opacity duration-150"
          :class="isRefreshing && !loading ? 'animate-pulse opacity-100' : 'opacity-0'"
        />
      </div>

      <StatePanel v-if="loading" message="Loading attendance records..." title="Please wait" />

      <StatePanel v-else-if="error" :message="error" title="Attendance load failed" tone="error" />

      <div v-else-if="records.length > 0" class="overflow-x-auto rounded-lg border border-subtle">
        <table class="w-full text-left text-sm text-on-surface">
          <thead class="bg-surface-variant/40 text-on-surface-variant">
            <tr>
              <th class="px-3 py-2">Name</th>
              <th class="px-3 py-2">Participants</th>
              <th class="px-3 py-2">Recorded</th>
              <th class="px-3 py-2">Date</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="record in records" :key="record.id" class="border-t border-subtle">
              <td class="px-3 py-2">{{ record.name }}</td>
              <td class="px-3 py-2">{{ record.participantCount }}</td>
              <td class="px-3 py-2">{{ record.recorded ? 'Yes' : 'No' }}</td>
              <td class="px-3 py-2">{{ new Date(record.dateCreated).toLocaleDateString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="text-sm text-on-surface-variant">
        {{ search ? 'No attendance records matched your search.' : 'No attendance records available.' }}
      </p>

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
