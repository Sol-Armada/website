<script setup lang="ts">
  import { Switch } from '@vuetify/v0'
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type AttendanceRecord, type MemberSummary } from '@/services/adminService'
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
  const isCreateModalOpen = ref(false)
  const creating = ref(false)
  const createError = ref<string | null>(null)
  const createSuccess = ref<string | null>(null)
  const availableAttendanceNames = ref<string[]>([])
  const eventNameSearch = ref('')
  const eventNameFocused = ref(false)
  const availableMembers = ref<Record<string, string>>({})
  const selectedParticipantIDs = ref<string[]>([])
  const memberSearch = ref('')
  const memberSearchFocused = ref(false)
  const memberSearchLoading = ref(false)
  const memberSearchResults = ref<MemberSummary[]>([])
  const createFormName = ref('')
  const createFormAllowTokens = ref(false)
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let memberSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let refreshTimer: number | null = null
  let inFlightRequest: Promise<void> | null = null
  let queuedRefreshMode: 'background' | 'blocking' | null = null
  const unsubscribers: Array<() => void> = []

  const filteredMemberResults = computed(() => {
    const selected = new Set(selectedParticipantIDs.value)
    return memberSearchResults.value.filter(member => !selected.has(member.id))
  })

  const selectedParticipants = computed(() => {
    return selectedParticipantIDs.value.map(id => ({
      id,
      name: availableMembers.value[id] || id,
    }))
  })

  const filteredAttendanceNames = computed(() => {
    const query = eventNameSearch.value.trim().toLowerCase()
    if (query === '') {
      return availableAttendanceNames.value
    }

    return availableAttendanceNames.value.filter(name => name.toLowerCase().includes(query))
  })

  function resetCreateForm(): void {
    createFormName.value = ''
    eventNameSearch.value = ''
    eventNameFocused.value = false
    memberSearch.value = ''
    selectedParticipantIDs.value = []
    createFormAllowTokens.value = false
  }

  function openCreateModal(): void {
    void Promise.all([
      adminService.getAvailableAttendanceNames(),
      adminService.getMembers(100, 1),
    ]).then(([names, membersResponse]) => {
      availableAttendanceNames.value = names
      memberSearchResults.value = membersResponse.members || []
      availableMembers.value = memberSearchResults.value.reduce<Record<string, string>>((acc, member) => {
        acc[member.id] = member.username
        return acc
      }, {})
    }).catch(error_ => {
      availableAttendanceNames.value = []
      memberSearchResults.value = []
      console.error('Failed to fetch modal options:', error_)
    })

    createError.value = null
    createSuccess.value = null
    isCreateModalOpen.value = true
  }

  function closeCreateModal(): void {
    if (creating.value) {
      return
    }
    isCreateModalOpen.value = false
    createError.value = null
    resetCreateForm()
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

  function addParticipant(member: MemberSummary): void {
    if (selectedParticipantIDs.value.includes(member.id)) {
      return
    }

    selectedParticipantIDs.value = [...selectedParticipantIDs.value, member.id]
    availableMembers.value = {
      ...availableMembers.value,
      [member.id]: member.username,
    }

    memberSearch.value = ''
    void searchMembers('')
  }

  function removeParticipant(memberID: string): void {
    selectedParticipantIDs.value = selectedParticipantIDs.value.filter(id => id !== memberID)
  }

  function selectEventName(name: string): void {
    createFormName.value = name
    eventNameSearch.value = name
    eventNameFocused.value = false
  }

  async function submitAttendanceRecord(): Promise<void> {
    if (creating.value) {
      return
    }

    createError.value = null
    createSuccess.value = null

    const name = createFormName.value.trim()
    const participantIdentifiers = [...selectedParticipantIDs.value]

    if (name.length === 0) {
      createError.value = 'Event name is required.'
      return
    }

    if (participantIdentifiers.length === 0) {
      createError.value = 'Select at least one participant.'
      return
    }

    creating.value = true
    try {
      await adminService.createAttendanceRecord({
        name,
        participantIdentifiers,
      })

      createSuccess.value = 'Attendance record created successfully.'
      isCreateModalOpen.value = false
      resetCreateForm()

      if (page.value === 1) {
        void loadAttendance({ background: true })
      } else {
        page.value = 1
      }
    } catch(error_: any) {
      const message = error_?.message || 'Failed to create attendance record'
      createError.value = message.includes('(404)')
        ? 'The create attendance endpoint is not available yet. Add POST /api/admin/attendance on the API to enable this action.'
        : message
    } finally {
      creating.value = false
    }
  }

  function handleGlobalKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape' && isCreateModalOpen.value) {
      closeCreateModal()
    }
  }

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

  watch(memberSearch, value => {
    if (memberSearchDebounceTimer) {
      clearTimeout(memberSearchDebounceTimer)
    }

    memberSearchDebounceTimer = setTimeout(() => {
      void searchMembers(value.trim())
    }, 250)
  })

  watch(isCreateModalOpen, value => {
    document.body.style.overflow = value ? 'hidden' : ''
  })

  onMounted(async() => {
    await loadAttendance()
    unsubscribers.push(wsClient.onTopic(WS_TOPIC_ADMIN_ATTENDANCE, scheduleRefresh))
    window.addEventListener('keydown', handleGlobalKeydown)
  })

  onBeforeUnmount(() => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }
    if (memberSearchDebounceTimer) {
      clearTimeout(memberSearchDebounceTimer)
    }
    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
      refreshTimer = null
    }
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
    window.removeEventListener('keydown', handleGlobalKeydown)
    document.body.style.overflow = ''
  })
</script>

<template>
  <PortalShell>
    <PageHeader subtitle="" title="Attendance" />

    <DataPanel description="Search attendance records and create new entries." title="Attendance Records">
      <StatePanel
        v-if="createSuccess"
        class="mb-3"
        :message="createSuccess"
        title="Attendance Saved"
        tone="success"
      />

      <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
        <input
          v-model="search"
          class="w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
          placeholder="Search attendance..."
          type="search"
        >

        <button
          class="shrink-0 rounded-md border border-primary bg-primary px-4 py-2 text-sm font-semibold text-on-primary transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          type="button"
          @click="openCreateModal"
        >
          + Record Attendance
        </button>
      </div>

      <div class="mb-3 mt-2 h-0.5 w-full overflow-hidden rounded-full bg-surface-variant/40">
        <div
          class="h-full w-full bg-primary/80 transition-opacity duration-150"
          :class="isRefreshing && !loading ? 'animate-pulse opacity-100' : 'opacity-0'"
        />
      </div>

      <StatePanel v-if="error" :message="error" title="Attendance load failed" tone="error" />

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

    <div
      v-if="isCreateModalOpen"
      class="attendance-modal-overlay"
      role="presentation"
      @click.self="closeCreateModal"
    >
      <div class="attendance-modal-panel">
        <div class="attendance-modal-header">
          <div>
            <h2 class="text-lg font-semibold text-on-surface">Record Attendance</h2>
            <p class="mt-1 text-sm text-on-surface-variant">Create a new attendance record with participant member identifiers.</p>
          </div>

          <button
            aria-label="Close attendance modal"
            class="rounded-md p-2 text-on-surface-variant transition hover:bg-surface-variant/40 hover:text-on-surface"
            type="button"
            @click="closeCreateModal"
          >
            x
          </button>
        </div>

        <form class="attendance-modal-body" @submit.prevent="submitAttendanceRecord">
          <StatePanel
            v-if="createError"
            class="mb-2"
            :message="createError"
            title="Unable To Save Attendance"
            tone="error"
          />

          <label class="text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="attendance-name">
            Event Name
          </label>

          <div class="attendance-event-picker">
            <input
              id="attendance-name"
              v-model="eventNameSearch"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              placeholder="Select an event name"
              type="search"
              @blur="eventNameFocused = false"
              @focus="eventNameFocused = true"
              @input="createFormName = ''"
            >

            <div v-if="eventNameFocused" class="attendance-event-menu">
              <p v-if="filteredAttendanceNames.length === 0" class="attendance-event-menu__status">
                No events found.
              </p>

              <button
                v-for="name in filteredAttendanceNames"
                v-else
                :key="name"
                class="attendance-event-menu__item"
                type="button"
                @mousedown.prevent="selectEventName(name)"
              >
                {{ name }}
              </button>
            </div>
          </div>

          <label class="mt-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="attendance-participants">
            Participants
          </label>

          <div class="attendance-participant-picker">
            <div v-if="selectedParticipants.length > 0" class="attendance-participant-chips mb-1">
              <button
                v-for="participant in selectedParticipants"
                :key="participant.id"
                class="attendance-participant-chip"
                type="button"
                @click="removeParticipant(participant.id)"
              >
                <span>{{ participant.name }}</span>
                <span class="attendance-participant-chip__remove">x</span>
              </button>
            </div>

            <input
              id="attendance-participants"
              v-model="memberSearch"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              placeholder="Search members by name..."
              type="search"
              @blur="memberSearchFocused = false"
              @focus="memberSearchFocused = true"
            >

            <div v-if="memberSearchFocused" class="attendance-member-menu">
              <p v-if="memberSearchLoading" class="attendance-member-menu__status">Searching members...</p>

              <p v-else-if="filteredMemberResults.length === 0" class="attendance-member-menu__status">
                No members found.
              </p>

              <button
                v-for="member in filteredMemberResults"
                v-else
                :key="member.id"
                class="attendance-member-menu__item"
                type="button"
                @mousedown.prevent="addParticipant(member)"
              >
                {{ member.username }}
              </button>
            </div>
          </div>

          <p class="text-xs text-on-surface-variant">
            Search and select members. Click a chip to remove it.
          </p>

          <div class="mt-4 flex items-center justify-between gap-3">
            <label class="text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="allow-tokens-switch">
              Award Tokens
            </label>

            <Switch.Root
              id="allow-tokens-switch"
              v-slot="{ isChecked }"
              v-model="createFormAllowTokens"
              class="attendance-switch-root"
            >
              <span class="attendance-switch-track" :class="{ 'attendance-switch-track--checked': isChecked }">
                <span class="attendance-switch-thumb" :class="{ 'attendance-switch-thumb--checked': isChecked }" />
              </span>
            </Switch.Root>
          </div>

          <div class="mt-2 flex items-center justify-end gap-2">
            <button
              class="rounded-md border border-subtle px-4 py-2 text-sm text-on-surface transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="creating"
              type="button"
              @click="closeCreateModal"
            >
              Cancel
            </button>

            <button
              class="rounded-md border border-primary bg-primary px-4 py-2 text-sm font-semibold text-on-primary transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="creating"
              type="submit"
            >
              {{ creating ? 'Saving...' : 'Save Attendance' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </PortalShell>
</template>

<style scoped>
  .attendance-modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: rgb(0 0 0 / 0.38);
    backdrop-filter: blur(4px);
  }

  .attendance-modal-panel {
    width: 100%;
    max-width: 40rem;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    background: var(--v0-surface);
    box-shadow: 0 24px 72px rgb(0 0 0 / 0.45);
    overflow: visible;
    position: relative;
  }

  .attendance-modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 1.25rem 1.25rem 1rem;
    border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
  }

  .attendance-modal-body {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 1rem 1.25rem 1.25rem;
  }

  .attendance-event-picker {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .attendance-event-menu {
    position: absolute;
    z-index: 60;
    top: calc(100% + 0.2rem);
    left: 0;
    right: 0;
    max-height: 16rem;
    overflow-y: auto;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    border-radius: 0.5rem;
    background: var(--v0-surface);
    box-shadow: 0 14px 30px rgb(0 0 0 / 0.3);
  }

  .attendance-event-menu__item,
  .attendance-event-menu__status {
    width: 100%;
    display: block;
    text-align: left;
    padding: 0.55rem 0.75rem;
    font-size: 0.875rem;
    color: var(--v0-on-surface);
  }

  .attendance-event-menu__status {
    color: var(--v0-on-surface-variant);
  }

  .attendance-event-menu__item:hover {
    background: color-mix(in srgb, var(--v0-primary) 12%, transparent);
  }

  .attendance-participant-picker {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .attendance-participant-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .attendance-participant-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    border: 1px solid color-mix(in srgb, var(--v0-primary) 40%, transparent);
    background: color-mix(in srgb, var(--v0-primary) 14%, transparent);
    color: var(--v0-on-surface);
    border-radius: 999px;
    padding: 0.2rem 0.65rem;
    font-size: 0.75rem;
    transition: background-color 150ms ease;
  }

  .attendance-participant-chip:hover {
    background: color-mix(in srgb, var(--v0-primary) 22%, transparent);
  }

  .attendance-participant-chip__remove {
    color: var(--v0-on-surface-variant);
    font-weight: 700;
    line-height: 1;
  }

  .attendance-member-menu {
    position: absolute;
    z-index: 60;
    top: calc(100% + 0.2rem);
    left: 0;
    right: 0;
    max-height: 16rem;
    overflow-y: auto;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    border-radius: 0.5rem;
    background: var(--v0-surface);
    box-shadow: 0 14px 30px rgb(0 0 0 / 0.3);
  }

  .attendance-member-menu__item,
  .attendance-member-menu__status {
    width: 100%;
    display: block;
    text-align: left;
    padding: 0.55rem 0.75rem;
    font-size: 0.875rem;
    color: var(--v0-on-surface);
  }

  .attendance-member-menu__status {
    color: var(--v0-on-surface-variant);
  }

  .attendance-member-menu__item:hover {
    background: color-mix(in srgb, var(--v0-primary) 12%, transparent);
  }

  :deep(.attendance-switch-root) {
    display: inline-flex;
    align-items: center;
    cursor: pointer;
    border-radius: 999px;
  }

  .attendance-switch-track {
    display: inline-flex;
    align-items: center;
    width: 46px;
    height: 26px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    background: color-mix(in srgb, var(--v0-surface) 84%, black 16%);
    padding: 2px;
    transition: background-color 150ms ease, border-color 150ms ease;
  }

  .attendance-switch-thumb {
    width: 20px;
    height: 20px;
    border-radius: 999px;
    background: var(--v0-on-surface-variant);
    transform: translateX(0);
    transition: transform 150ms ease, background-color 150ms ease;
  }

  .attendance-switch-track--checked {
    background: color-mix(in srgb, var(--v0-primary) 32%, transparent);
    border-color: color-mix(in srgb, var(--v0-primary) 65%, transparent);
  }

  .attendance-switch-thumb--checked {
    transform: translateX(20px);
    background: var(--v0-primary);
  }

  @media (max-width: 720px) {
    .attendance-modal-overlay {
      align-items: flex-end;
      padding: 0;
    }

    .attendance-modal-panel {
      max-width: none;
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
      max-height: 90vh;
      overflow-y: auto;
    }
  }
</style>
