<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type MemberSummary } from '@/services/adminService'

  const route = useRoute()
  const router = useRouter()
  const routeId = (route.params as Record<string, string | string[] | undefined>).id
  const attendanceId = Array.isArray(routeId) ? routeId[0] : routeId

  const isLoading = ref(true)
  const isSaving = ref(false)
  const error = ref<string | null>(null)
  const saveError = ref<string | null>(null)

  const availableAttendanceNames = ref<string[]>([])
  const availableMembers = ref<Record<string, string>>({})

  const name = ref('')
  const recorded = ref(false)
  const successful = ref(false)
  const awardTokens = ref(false)
  const participants = ref<MemberSummary[]>([])

  const memberSearch = ref('')
  const memberSearchFocused = ref(false)
  const memberSearchLoading = ref(false)
  const memberSearchResults = ref<MemberSummary[]>([])
  let memberSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null

  const selectedParticipantIDs = computed(() => new Set(participants.value.map(participant => participant.id)))

  const filteredMemberResults = computed(() => {
    return memberSearchResults.value.filter(member => !selectedParticipantIDs.value.has(member.id))
  })

  function getInitials(label: string): string {
    return label
      .split(' ')
      .map(part => part[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  function goBack(): void {
    if (window.history.length > 1) {
      router.back()
      return
    }

    if (!attendanceId) {
      void router.push('/admin/attendance')
      return
    }

    void router.push(`/admin/attendance/${attendanceId}`)
  }

  function addParticipant(member: MemberSummary): void {
    if (selectedParticipantIDs.value.has(member.id)) {
      return
    }

    availableMembers.value = {
      ...availableMembers.value,
      [member.id]: member.username,
    }

    participants.value = [
      ...participants.value,
      {
        ...member,
        onTime: false,
      },
    ]

    memberSearch.value = ''
  }

  function removeParticipant(memberID: string): void {
    participants.value = participants.value.filter(member => member.id !== memberID)
  }

  function hideMemberSearchMenu(): void {
    window.setTimeout(() => {
      memberSearchFocused.value = false
    }, 120)
  }

  async function searchMembers(query: string): Promise<void> {
    if (query.trim() === '') {
      memberSearchResults.value = []
      return
    }

    memberSearchLoading.value = true
    try {
      const response = await adminService.getMembers(100, 1, query.trim())
      const members = response.members || []
      memberSearchResults.value = members

      const nextMap = { ...availableMembers.value }
      for (const member of members) {
        nextMap[member.id] = member.username
      }
      availableMembers.value = nextMap
    } catch {
      memberSearchResults.value = []
    } finally {
      memberSearchLoading.value = false
    }
  }

  async function loadPage(): Promise<void> {
    if (!attendanceId) {
      error.value = 'Missing attendance ID'
      isLoading.value = false
      return
    }

    isLoading.value = true
    error.value = null

    try {
      const [payload, names] = await Promise.all([
        adminService.getAttendanceEditPayload(attendanceId),
        adminService.getAvailableAttendanceNames(),
      ])

      availableAttendanceNames.value = names
      name.value = payload.record.name
      recorded.value = payload.record.recorded
      successful.value = payload.record.successful
      participants.value = payload.participants.map(participant => ({
        ...participant,
        onTime: participant.onTime === true,
      }))

      const availableByID: Record<string, string> = {}
      for (const participant of payload.participants) {
        availableByID[participant.id] = participant.username
      }
      availableMembers.value = availableByID
    } catch(error_) {
      error.value = error_ instanceof Error ? error_.message : 'Failed to load attendance edit payload'
    } finally {
      isLoading.value = false
    }
  }

  async function submitForm(): Promise<void> {
    if (!attendanceId || isSaving.value) {
      return
    }

    saveError.value = null

    const normalizedName = name.value.trim()
    if (normalizedName === '') {
      saveError.value = 'Event name is required.'
      return
    }

    if (!availableAttendanceNames.value.includes(normalizedName)) {
      saveError.value = 'Event name must be selected from the approved list.'
      return
    }

    if (participants.value.length === 0) {
      saveError.value = 'At least one participant is required.'
      return
    }

    isSaving.value = true
    try {
      await adminService.updateAttendanceRecord(attendanceId, {
        name: normalizedName,
        recorded: recorded.value,
        successful: successful.value,
        awardTokens: awardTokens.value,
        participantIds: participants.value.map(participant => participant.id),
        onTimeParticipantIds: participants.value.filter(participant => participant.onTime === true).map(participant => participant.id),
      })

      void router.push(`/admin/attendance/${attendanceId}`)
    } catch(error_) {
      saveError.value = error_ instanceof Error ? error_.message : 'Failed to save attendance record'
    } finally {
      isSaving.value = false
    }
  }

  watch(memberSearch, query => {
    if (memberSearchDebounceTimer !== null) {
      clearTimeout(memberSearchDebounceTimer)
    }

    memberSearchDebounceTimer = setTimeout(() => {
      void searchMembers(query)
    }, 250)
  })

  onMounted(() => {
    void loadPage()
  })
</script>

<template>
  <PortalShell>
    <div class="attendance-edit">
      <div v-if="isLoading" class="loading-state">
        Loading attendance record...
      </div>

      <StatePanel
        v-else-if="error"
        :message="error"
        title="Unable To Load Attendance Record"
        tone="error"
      />

      <div v-else>
        <a class="back-link" href="#" @click.prevent="goBack">
          ← Back to Record Details
        </a>

        <PageHeader
          subtitle="Update event details, manage participants, and adjust on-time tracking."
          title="Edit Attendance Record"
        />

        <form @submit.prevent="submitForm">
          <DataPanel
            description="Choose the event name and attendance status values."
            title="Event Details"
          >
            <div class="event-grid">
              <label class="field-group">
                <span class="field-label">Event Name</span>

                <select v-model="name" class="field-control" required>
                  <option disabled value="">Select an event name</option>

                  <option v-for="attendanceName in availableAttendanceNames" :key="attendanceName" :value="attendanceName">
                    {{ attendanceName }}
                  </option>
                </select>
              </label>

              <label class="field-group checkbox-group">
                <input v-model="recorded" class="checkbox-control" type="checkbox">
                <span>Recorded</span>
              </label>

              <label class="field-group checkbox-group">
                <input v-model="successful" class="checkbox-control" type="checkbox">
                <span>Successful</span>
              </label>

              <label class="field-group checkbox-group">
                <input v-model="awardTokens" class="checkbox-control" type="checkbox">
                <span>Award Tokens</span>
              </label>
            </div>
          </DataPanel>

          <DataPanel
            :description="`Participants (${participants.length})`"
            title="Participants"
          >
            <div class="add-participants-row">
              <div class="member-search-wrap">
                <input
                  v-model="memberSearch"
                  class="field-control"
                  placeholder="Search members to add..."
                  type="text"
                  @blur="hideMemberSearchMenu"
                  @focus="memberSearchFocused = true"
                >

                <div v-if="memberSearchFocused && memberSearch.trim() !== ''" class="search-menu">
                  <div v-if="memberSearchLoading" class="search-placeholder">
                    Searching...
                  </div>

                  <button
                    v-for="member in filteredMemberResults"
                    :key="member.id"
                    class="search-item"
                    type="button"
                    @mousedown.prevent="addParticipant(member)"
                  >
                    <span class="search-name">{{ member.username }}</span>
                    <span class="search-meta">{{ member.rank }}</span>
                  </button>

                  <div v-if="!memberSearchLoading && filteredMemberResults.length === 0" class="search-placeholder">
                    No members found.
                  </div>
                </div>
              </div>
            </div>

            <div v-if="participants.length === 0" class="empty-message">
              Add at least one participant to continue.
            </div>

            <div v-else class="participants-grid">
              <article v-for="participant in participants" :key="participant.id" class="participant-card">
                <div class="participant-avatar">
                  {{ getInitials(participant.username) }}
                </div>

                <div class="participant-info">
                  <p class="participant-name">{{ participant.username }}</p>
                  <p class="participant-rank">{{ participant.rank }}</p>
                </div>

                <button
                  :aria-pressed="participant.onTime === true"
                  class="on-time-button"
                  :class="{ 'on-time-button-active': participant.onTime }"
                  type="button"
                  @click="participant.onTime = !participant.onTime"
                >
                  {{ participant.onTime ? 'On Time' : 'Not On Time' }}
                </button>

                <button
                  class="remove-button"
                  type="button"
                  @click="removeParticipant(participant.id)"
                >
                  ×
                </button>
              </article>
            </div>
          </DataPanel>

          <StatePanel
            v-if="saveError"
            :message="saveError"
            title="Unable To Save"
            tone="error"
          />

          <div class="actions-row">
            <button class="save-button" :disabled="isSaving" type="submit">
              {{ isSaving ? 'Saving...' : 'Save Changes' }}
            </button>

            <button class="cancel-button" :disabled="isSaving" type="button" @click="goBack">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  </PortalShell>
</template>

<style scoped>
  .attendance-edit {
    max-width: 1180px;
    margin: 0 auto;
    padding: 2.25rem 1.5rem;
  }

  .loading-state {
    padding: 2rem;
    border: 1px solid var(--v0-divider);
    border-radius: 1rem;
    color: var(--v0-on-surface-variant);
    background: var(--v0-surface);
  }

  .back-link {
    display: inline-flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
    color: var(--v0-on-surface-variant);
    text-decoration: none;
    font-size: 0.875rem;
  }

  .back-link:hover {
    color: var(--v0-primary);
  }

  .event-grid {
    display: grid;
    grid-template-columns: minmax(240px, 1fr) repeat(2, minmax(160px, auto));
    gap: 1rem;
    align-items: center;
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .field-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--v0-on-surface-variant);
  }

  .field-control {
    width: 100%;
    border: 1px solid var(--v0-divider);
    border-radius: 0.75rem;
    background: color-mix(in oklab, var(--v0-surface), black 6%);
    color: var(--v0-on-surface);
    padding: 0.65rem 0.8rem;
    outline: none;
  }

  .field-control:focus {
    border-color: var(--v0-primary);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--v0-primary), transparent 82%);
  }

  .checkbox-group {
    flex-direction: row;
    align-items: center;
    gap: 0.5rem;
    padding-top: 1rem;
  }

  .checkbox-control {
    width: 1rem;
    height: 1rem;
    accent-color: var(--v0-primary);
  }

  .add-participants-row {
    margin-bottom: 1rem;
  }

  .member-search-wrap {
    position: relative;
    max-width: 420px;
  }

  .search-menu {
    position: absolute;
    z-index: 20;
    top: calc(100% + 0.4rem);
    width: 100%;
    max-height: 16rem;
    overflow-y: auto;
    border: 1px solid var(--v0-divider);
    border-radius: 0.85rem;
    background: color-mix(in oklab, var(--v0-surface), black 4%);
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.26);
  }

  .search-item {
    width: 100%;
    border: 0;
    border-bottom: 1px solid color-mix(in oklab, var(--v0-divider), transparent 35%);
    padding: 0.7rem 0.85rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: transparent;
    color: var(--v0-on-surface);
    text-align: left;
    cursor: pointer;
  }

  .search-item:hover {
    background: color-mix(in oklab, var(--v0-primary), transparent 88%);
  }

  .search-name {
    font-weight: 600;
  }

  .search-meta {
    font-size: 0.8rem;
    color: var(--v0-on-surface-variant);
  }

  .search-placeholder {
    padding: 0.75rem 0.85rem;
    font-size: 0.875rem;
    color: var(--v0-on-surface-variant);
  }

  .empty-message {
    border: 1px dashed var(--v0-divider);
    border-radius: 0.9rem;
    padding: 1rem;
    color: var(--v0-on-surface-variant);
  }

  .participants-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
    gap: 0.9rem;
  }

  .participant-card {
    border: 1px solid var(--v0-divider);
    border-radius: 0.9rem;
    background: color-mix(in oklab, var(--v0-surface), black 4%);
    padding: 0.75rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .participant-avatar {
    width: 2.35rem;
    height: 2.35rem;
    border-radius: 999px;
    display: grid;
    place-items: center;
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.03em;
    color: #0a1220;
    background: linear-gradient(140deg, #E6A82D 0%, #0041FF 100%);
  }

  .participant-info {
    min-width: 0;
    flex: 1;
  }

  .participant-name {
    margin: 0;
    font-weight: 600;
    font-size: 0.95rem;
    color: var(--v0-on-surface);
  }

  .participant-rank {
    margin: 0;
    font-size: 0.75rem;
    color: var(--v0-on-surface-variant);
  }

  .on-time-button {
    border: 1px solid var(--v0-divider);
    background: color-mix(in oklab, var(--v0-surface), black 4%);
    color: var(--v0-on-surface-variant);
    border-radius: 999px;
    padding: 0.3rem 0.6rem;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    cursor: pointer;
    transition: all var(--motion-fast) var(--ease-standard);
    white-space: nowrap;
  }

  .on-time-button:hover {
    border-color: color-mix(in oklab, var(--v0-primary), var(--v0-divider) 45%);
    color: var(--v0-on-surface);
  }

  .on-time-button-active {
    border-color: color-mix(in oklab, #22c55e, transparent 62%);
    background: color-mix(in oklab, #22c55e, transparent 86%);
    color: #86efac;
  }

  .remove-button {
    border: 1px solid var(--v0-divider);
    background: transparent;
    color: var(--v0-on-surface-variant);
    width: 1.85rem;
    height: 1.85rem;
    border-radius: 0.55rem;
    font-size: 1rem;
    line-height: 1;
    cursor: pointer;
  }

  .remove-button:hover {
    border-color: #fb7185;
    color: #fb7185;
  }

  .actions-row {
    display: flex;
    gap: 0.75rem;
    margin-top: 1.25rem;
  }

  .save-button,
  .cancel-button {
    padding: 0.65rem 1rem;
    border-radius: 0.7rem;
    border: 1px solid transparent;
    cursor: pointer;
    font-weight: 600;
  }

  .save-button {
    background: var(--v0-primary);
    color: var(--v0-on-primary);
  }

  .save-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .cancel-button {
    background: transparent;
    border-color: var(--v0-divider);
    color: var(--v0-on-surface);
  }

  @media (max-width: 860px) {
    .event-grid {
      grid-template-columns: 1fr;
      align-items: start;
    }

    .checkbox-group {
      padding-top: 0;
    }

    .actions-row {
      flex-direction: column;
    }

    .save-button,
    .cancel-button {
      width: 100%;
    }
  }
</style>
