<script setup lang="ts">
  import { Switch } from '@vuetify/v0'
  import { storeToRefs } from 'pinia'
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type MemberSummary } from '@/services/adminService'
  import { useAttendanceStore } from '@/stores/attendance'
  import { useAuthStore } from '@/stores/auth'

  const router = useRouter()
  const authStore = useAuthStore()
  const attendanceStore = useAttendanceStore()
  const {
    loading,
    isRefreshing,
    error,
    records,
    search,
    page,
    pageInput,
    hasNextPage,
    analyticsLoading,
    analyticsRefreshing,
    analyticsError,
    attendanceAnalytics,
    availableAttendanceNames,
    availableMembers,
    memberSearchLoading,
    memberSearchResults,
    managerSearchLoading,
    managerSearchResults,
  } = storeToRefs(attendanceStore)

  const isCreateModalOpen = ref(false)
  const creating = ref(false)
  const createError = ref<string | null>(null)
  const createSuccess = ref<string | null>(null)
  const eventNameSearch = ref('')
  const eventNameFocused = ref(false)
  const selectedParticipantIDs = ref<string[]>([])
  const memberSearch = ref('')
  const memberSearchFocused = ref(false)
  const selectedManagerIDs = ref<string[]>([])
  const managerSearch = ref('')
  const managerSearchFocused = ref(false)
  const createFormName = ref('')
  const createFormAllowTokens = ref(false)
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let memberSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let managerSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  const currentUserId = computed(() => authStore.user?.id ?? null)

  const filteredMemberResults = computed(() => {
    const selected = new Set(selectedParticipantIDs.value)
    return memberSearchResults.value.filter(member => !selected.has(member.id))
  })

  const filteredManagerResults = computed(() => {
    const selected = new Set(selectedManagerIDs.value)
    return managerSearchResults.value.filter(manager => !selected.has(manager.id))
  })

  const selectedParticipants = computed(() => {
    return selectedParticipantIDs.value.map(id => ({
      id,
      name: availableMembers.value[id] || id,
    }))
  })

  const selectedManagers = computed(() => {
    return selectedManagerIDs.value.map(id => ({
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

  function formatAnalyticsWindowLabel(start: string, end: string): string {
    const startLabel = new Date(start).toLocaleDateString()
    const endLabel = new Date(end).toLocaleDateString()
    return `${startLabel} - ${endLabel}`
  }

  function resetCreateForm(): void {
    createFormName.value = ''
    eventNameSearch.value = ''
    eventNameFocused.value = false
    memberSearch.value = ''
    selectedParticipantIDs.value = []
    selectedManagerIDs.value = []
    createFormAllowTokens.value = false
  }

  function openCreateModal(): void {
    void attendanceStore.loadCreateModalOptions()

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
    await attendanceStore.searchMembers(query)
  }

  async function searchManagers(query: string): Promise<void> {
    await attendanceStore.searchManagers(query)
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

  function addManager(member: MemberSummary): void {
    if (selectedManagerIDs.value.includes(member.id)) {
      return
    }

    selectedManagerIDs.value = [...selectedManagerIDs.value, member.id]
    availableMembers.value = {
      ...availableMembers.value,
      [member.id]: member.username,
    }

    addParticipant(member)

    managerSearch.value = ''
    void searchManagers('')
  }

  function removeParticipant(memberID: string): void {
    selectedParticipantIDs.value = selectedParticipantIDs.value.filter(id => id !== memberID)
    selectedManagerIDs.value = selectedManagerIDs.value.filter(id => id !== memberID)
  }

  function removeManager(memberID: string): void {
    selectedManagerIDs.value = selectedManagerIDs.value.filter(id => id !== memberID)
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
    const participantIds = [...selectedParticipantIDs.value]
    const managerIds = [...selectedManagerIDs.value]
    const awardTokens = createFormAllowTokens.value

    if (name.length === 0) {
      createError.value = 'Event name is required.'
      return
    }

    if (participantIds.length === 0) {
      createError.value = 'Select at least one participant.'
      return
    }

    creating.value = true
    try {
      await adminService.createAttendanceRecord({
        name,
        participantIds,
        managerIds,
        awardTokens,
        submittedBy: currentUserId.value || null,
      })

      createSuccess.value = 'Attendance record created successfully.'
      isCreateModalOpen.value = false
      resetCreateForm()

      if (page.value === 1) {
        void attendanceStore.loadAttendance({ background: true })
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
    void attendanceStore.loadAttendance()
  })

  watch(search, () => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }

    searchDebounceTimer = setTimeout(() => {
      page.value = 1
      void attendanceStore.loadAttendance({ background: true })
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

  watch(managerSearch, value => {
    if (managerSearchDebounceTimer) {
      clearTimeout(managerSearchDebounceTimer)
    }

    managerSearchDebounceTimer = setTimeout(() => {
      void searchManagers(value.trim())
    }, 250)
  })

  watch(isCreateModalOpen, value => {
    document.body.style.overflow = value ? 'hidden' : ''
  })

  onMounted(async() => {
    await attendanceStore.initialize()
    window.addEventListener('keydown', handleGlobalKeydown)
  })

  onBeforeUnmount(() => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }
    if (memberSearchDebounceTimer) {
      clearTimeout(memberSearchDebounceTimer)
    }
    if (managerSearchDebounceTimer) {
      clearTimeout(managerSearchDebounceTimer)
    }
    attendanceStore.dispose()
    window.removeEventListener('keydown', handleGlobalKeydown)
    document.body.style.overflow = ''
  })
</script>

<template>
  <PortalShell>
    <PageHeader subtitle="" title="Attendance Records" />

    <DataPanel description="" title="Analytics">
      <p
        v-if="analyticsRefreshing && !analyticsLoading"
        class="mb-3 text-xs font-medium uppercase tracking-wide text-on-surface-variant"
      >
        Refreshing analytics...
      </p>

      <StatePanel v-if="analyticsLoading" message="Loading attendance analytics..." title="Please wait" />

      <StatePanel v-else-if="analyticsError" :message="analyticsError" title="Analytics load failed" tone="error" />

      <div v-else-if="attendanceAnalytics" class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <StatCard
          :detail="
            attendanceAnalytics &&
              attendanceAnalytics.uniqueAttendeesLast30Days + attendanceAnalytics.inactiveMembersLast30Days > 0
              ? `${(
                attendanceAnalytics.uniqueAttendeesLast30Days /
                (attendanceAnalytics.uniqueAttendeesLast30Days + attendanceAnalytics.inactiveMembersLast30Days) *
                100
              ).toFixed(1)}%` + ' of members attended in the last 30 days'
              : '0%'
          "
          label="Unique Attendees (30 Days)"
          :value="attendanceAnalytics.uniqueAttendeesLast30Days"
        />

        <StatCard
          :detail="`${attendanceAnalytics.mostPopularEventAttendanceLast30Days} total attendees across ${attendanceAnalytics.totalEventsLast30Days} events`"
          label="Most Popular Event (30 Days)"
          :value="attendanceAnalytics.mostPopularEventLast30Days"
        />
      </div>
    </DataPanel>

    <DataPanel description="" title="">
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
              <th class="px-3 py-2">Award Tokens</th>
              <th class="px-3 py-2">Recorded</th>
              <th class="px-3 py-2">Date</th>
              <th class="px-3 py-2">Actions</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="record in records" :key="record.id" class="border-t border-subtle">
              <td class="px-3 py-2">{{ record.name }}</td>
              <td class="px-3 py-2">{{ record.participantCount }}</td>
              <td class="px-3 py-2">{{ record.awardTokens ? 'Yes' : 'No' }}</td>
              <td class="px-3 py-2">{{ record.recorded ? 'Yes' : 'No' }}</td>
              <td class="px-3 py-2">{{ new Date(record.dateCreated).toLocaleDateString() }}</td>

              <td class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <button
                    class="inline-flex items-center gap-1 px-3 py-1 bg-primary text-on-primary rounded-md text-sm font-medium hover:opacity-90 transition-opacity"
                    type="button"
                    @click="router.push(`/admin/attendance/${record.id}`)"
                  >
                    <i class="mdi mdi-eye" />
                    View
                  </button>

                  <button
                    class="inline-flex items-center gap-1 rounded-md border border-subtle px-3 py-1 text-sm font-medium text-on-surface transition hover:bg-surface-variant/40"
                    type="button"
                    @click="router.push(`/admin/attendance/${record.id}/edit`)"
                  >
                    <i class="mdi mdi-pencil" />
                    Edit
                  </button>
                </div>
              </td></tr>
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

          <label class="mt-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="attendance-participants">
            Event Managers
          </label>

          <div class="attendance-participant-picker">
            <div v-if="selectedManagers.length > 0" class="attendance-participant-chips mb-1">
              <button
                v-for="manager in selectedManagers"
                :key="manager.id"
                class="attendance-participant-chip"
                type="button"
                @click="removeManager(manager.id)"
              >
                <span>{{ manager.name }}</span>
                <span class="attendance-participant-chip__remove">x</span>
              </button>
            </div>

            <input
              id="attendance-managers"
              v-model="managerSearch"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              placeholder="Search members by name..."
              type="search"
              @blur="managerSearchFocused = false"
              @focus="managerSearchFocused = true"
            >

            <div v-if="managerSearchFocused" class="attendance-member-menu">
              <p v-if="managerSearchLoading" class="attendance-member-menu__status">Searching members...</p>

              <p v-else-if="filteredManagerResults.length === 0" class="attendance-member-menu__status">
                No members found.
              </p>

              <button
                v-for="manager in filteredManagerResults"
                v-else
                :key="manager.id"
                class="attendance-member-menu__item"
                type="button"
                @mousedown.prevent="addManager(manager)"
              >
                {{ manager.username }}
              </button>
            </div>
          </div>

          <p class="text-xs text-on-surface-variant">
            Search and select members. Click a chip to remove it.
          </p>

          <div class="mt-4 inline-flex items-center gap-3">
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
