<script setup lang="ts">
  import type { AttendanceRecord, MemberSummary } from '@/services/adminService'
  import { Popover } from '@vuetify/v0'
  import { onMounted, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import { adminService } from '@/services/adminService'

  const route = useRoute()
  const router = useRouter()
  const routeId = (route.params as Record<string, string | string[] | undefined>).id
  const attendanceId = Array.isArray(routeId) ? routeId[0] : routeId

  const attendanceRecord = ref<AttendanceRecord | null>(null)
  const participants = ref<MemberSummary[]>([])
  const isLoading = ref(true)
  const error = ref<string | null>(null)
  const participantsError = ref<string | null>(null)
  const hoveredParticipantId = ref<string | null>(null)

  onMounted(async() => {
    try {
      if (!attendanceId) {
        error.value = 'Missing attendance ID'
        return
      }

      attendanceRecord.value = await adminService.getAttendanceRecord(attendanceId)

      try {
        participants.value = await adminService.getMembersByAttendance(attendanceId)
      } catch(error_) {
        participantsError.value = error_ instanceof Error ? error_.message : 'Failed to load participants'
      }
    } catch(error_) {
      error.value = error_ instanceof Error ? error_.message : 'Failed to load attendance record'
    } finally {
      isLoading.value = false
    }
  })

  function formatDate(dateString: string) {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  }

  function getInitials(name: string) {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  function goBack() {
    router.push('/admin/attendance')
  }

  function goToEdit() {
    if (!attendanceId) {
      return
    }

    router.push(`/admin/attendance/${attendanceId}/edit`)
  }

  function toggleAwardTokens() {
    if (!attendanceId) {
      return
    }

    adminService.updateAttendanceRecord(attendanceId, {
      name: attendanceRecord.value?.name || '',
      recorded: attendanceRecord.value?.recorded || false,
      successful: attendanceRecord.value?.successful || false,
      awardTokens: !(attendanceRecord.value?.awardTokens || false),
      participantIds: participants.value.map(participant => participant.id),
      onTimeParticipantIds: participants.value.filter(participant => participant.onTime === true).map(participant => participant.id),
    }).then(payload => {
      attendanceRecord.value = payload.record
    }).catch(error_ => {
      error.value = error_ instanceof Error ? error_.message : 'Failed to update attendance record'
    })
  }

  function showOnTimePopover(participantId: string) {
    hoveredParticipantId.value = participantId
  }

  function hideOnTimePopover(participantId: string) {
    if (hoveredParticipantId.value === participantId) {
      hoveredParticipantId.value = null
    }
  }
</script>

<template>
  <PortalShell>
    <div class="attendance-detail">
      <div v-if="isLoading" class="loading">
        Loading attendance record...
      </div>

      <div v-else-if="error" class="error-container">
        <p>{{ error }}</p>
      </div>

      <div v-else-if="attendanceRecord" class="record-content">
        <a class="back-link" href="#" @click.prevent="goBack">
          ← Back to Attendance Reports
        </a>

        <div class="page-header">
          <h1 class="page-title">{{ attendanceRecord.name }}</h1>
        </div>

        <!-- Event Header -->
        <div class="event-header">
          <div class="event-meta">
            <div class="meta-item">
              <div class="meta-label">Submitted By</div>
              <div class="meta-value">{{ attendanceRecord.submittedBy }}</div>
            </div>

            <div class="meta-item">
              <div class="meta-label">Total Participants</div>
              <div class="meta-value">{{ participants.length || attendanceRecord.participantCount }}</div>
            </div>

            <div class="meta-item">
              <div class="meta-label">Status</div>

              <div class="meta-value">
                <span class="status-badge" :class="attendanceRecord.successful ? 'status-recorded' : 'status-pending'">
                  {{ attendanceRecord.recorded ? '✓ Recorded' : '◇ Pending' }}
                </span>
              </div>
            </div>

            <div class="meta-item">
              <div class="meta-label">Created</div>
              <div class="meta-value">{{ formatDate(attendanceRecord.dateCreated) }}</div>
            </div>

            <div class="meta-item">
              <div class="meta-label">Award Tokens</div>
              <div class="meta-value">{{ attendanceRecord.awardTokens ? 'Yes' : 'No' }}</div>
            </div>
          </div>

          <div class="action-bar">
            <button class="btn btn-primary" type="button" @click="goToEdit">Edit Record</button>
            <button class="btn btn-primary" type="button" @click="toggleAwardTokens">{{ attendanceRecord.awardTokens ? 'Don\'t Award Tokens' : 'Award Tokens' }}</button>
          </div>
        </div>

        <div class="participants-section">
          <h2 class="section-title">Participants ({{ participants.length || attendanceRecord.participantCount }})</h2>

          <p v-if="participantsError" class="participants-error">
            {{ participantsError }}
          </p>

          <p v-else-if="participants.length === 0" class="participants-empty">
            No participant list was returned for this attendance record.
          </p>

          <div v-else class="participants-grid">
            <div
              v-for="participant in participants"
              :key="participant.id"
              class="participant-card"
              :class="{ 'participant-card-on-time': participant.onTime }"
            >
              <div class="participant-avatar">
                <img v-if="participant.profileImage" :alt="participant.username" :src="participant.profileImage">
                <span v-else>{{ getInitials(participant.username) }}</span>
              </div>

              <div class="participant-info">
                <div class="participant-name">{{ participant.username }}</div>
                <div class="participant-rank">{{ participant.rank }}</div>

                <div v-if="attendanceRecord.awardTokens">
                  <Popover.Root :model-value="hoveredParticipantId === participant.id">
                    <Popover.Activator
                      as="span"
                      class="on-time-icon-wrap"
                      tabindex="0"
                      @blur="hideOnTimePopover(participant.id)"
                      @focus="showOnTimePopover(participant.id)"
                      @mouseenter="showOnTimePopover(participant.id)"
                      @mouseleave="hideOnTimePopover(participant.id)"
                    >
                      <i v-if="participant.onTime" class="mdi mdi-clock text-green" />
                      <i v-else class="mdi mdi-clock-outline text-red" />
                    </Popover.Activator>

                    <Popover.Content class="on-time-tooltip" position-area="top center" position-try="flip-block">
                      {{ participant.onTime ? 'On time' : 'Not on time' }}
                    </Popover.Content>
                  </Popover.Root>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="not-found">
        Attendance record not found
      </div>
    </div>
  </PortalShell>
</template>

<style scoped>
.attendance-detail {
  --bg: #090b12;
  --surface: #121722;
  --surface-warm: #1b2233;
  --fg: #f8fafc;
  --fg-2: #cbd5e1;
  --muted: #94a3b8;
  --border: #2a3447;
  --accent: #E6A82D;
  --accent-blue: #0041FF;
  --accent-on: #06101d;
  --accent-hover: #f0b841;
  --success: #22c55e;
  --danger: #fb7185;
  --radius-md: 16px;
  --motion-fast: 150ms;
  --ease-standard: cubic-bezier(0.2, 0, 0, 1);
  --text-xs: 12px;
  --text-sm: 14px;
  --text-base: 16px;
  --text-lg: 18px;
  --text-xl: 24px;
  --text-2xl: 36px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;

  max-width: 1200px;
  margin: 0 auto;
  padding: var(--space-8) var(--space-6);
  min-height: calc(100vh - 90px);
  background: var(--bg);
  color: var(--fg);
  position: relative;
  overflow: hidden;
  z-index: 1;
}

.attendance-detail::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(90deg, rgba(230, 168, 45, 0.03) 1px, transparent 1px),
    linear-gradient(0deg, rgba(230, 168, 45, 0.03) 1px, transparent 1px);
  background-size: 24px 24px;
  pointer-events: none;
  z-index: 0;
}

.record-content,
.loading,
.error-container,
.not-found {
  position: relative;
  z-index: 1;
}

.loading,
.error-container,
.not-found {
  padding: 2rem;
  text-align: center;
  font-size: 1.1rem;
  color: var(--fg-2);
}

.error-container {
  background: rgba(251, 113, 133, 0.1);
  border: 1px solid var(--danger);
  border-radius: var(--radius-md);
  color: var(--danger);
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--accent);
  text-decoration: none;
  font-size: var(--text-sm);
  margin-bottom: var(--space-6);
  transition: all var(--motion-fast) var(--ease-standard);
}

.back-link:hover {
  gap: var(--space-3);
  text-decoration: underline;
}

.page-header {
  margin-bottom: var(--space-8);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--accent);
  margin: 0;
  letter-spacing: -0.025em;
}

.event-header {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: var(--space-8);
  margin-bottom: var(--space-6);
  position: relative;
  overflow: hidden;
}

.event-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--accent) 0%, var(--accent-blue) 50%, transparent 100%);
  opacity: 0.8;
}

.event-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-6);
  margin-bottom: var(--space-6);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.meta-label {
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted);
  font-family: monospace;
}

.meta-value {
  font-size: var(--text-lg);
  font-weight: 500;
  color: var(--fg);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: 999px;
  font-size: var(--text-xs);
  font-family: monospace;
  font-weight: 500;
  width: fit-content;
}

.status-recorded {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: var(--success);
}

.status-pending {
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  color: #fbbf24;
}

.action-bar {
  display: flex;
  gap: var(--space-3);
  padding-top: var(--space-6);
  border-top: 1px solid var(--border);
}

.btn {
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--motion-fast) var(--ease-standard);
  border: none;
  font-family: inherit;
}

.btn-primary {
  background: var(--accent);
  color: var(--accent-on);
  border: 1px solid transparent;
}

.btn-primary:hover {
  background: var(--accent-hover);
  box-shadow: 0 0 24px rgba(230, 168, 45, 0.4);
}

.btn-secondary {
  background: transparent;
  color: var(--fg);
  border: 1px solid var(--border);
}

.btn-secondary:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.btn-danger {
  background: transparent;
  color: var(--danger);
  border: 1px solid var(--danger);
}

.btn-danger:hover {
  background: rgba(251, 113, 133, 0.1);
}

.participants-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: var(--space-8);
  margin-bottom: var(--space-6);
  position: relative;
  overflow: hidden;
}

.participants-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--accent) 0%, var(--accent-blue) 50%, transparent 100%);
  opacity: 0.8;
}

.section-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--fg);
  margin-bottom: var(--space-6);
}

.participants-error,
.participants-empty {
  color: var(--fg-2);
  font-size: var(--text-sm);
}

.participants-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-4);
}

.participant-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4);
  background: var(--surface-warm);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  transition: all var(--motion-fast) var(--ease-standard);
}

.participant-card-on-time {
  border-color: color-mix(in oklab, #22c55e, var(--border) 50%);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, #22c55e, transparent 72%);
}

.participant-card:hover {
  background: color-mix(in srgb, var(--accent) 5%, var(--surface-warm));
  border-color: var(--accent);
}

.participant-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--accent-on);
  flex-shrink: 0;
}

.participant-info {
  flex: 1;
  min-width: 0;
}

.participant-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--fg);
  margin-bottom: 2px;
}

.participant-rank {
  font-size: var(--text-xs);
  color: var(--muted);
  font-family: monospace;
}

.on-time-icon-wrap {
  display: inline-flex;
  align-items: center;
  margin-top: 4px;
}

.on-time-tooltip {
  z-index: 20;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--fg);
  padding: 4px 8px;
  font-size: 12px;
  line-height: 1.2;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
}

.on-time-badge {
  display: inline-flex;
  align-items: center;
  margin-top: 4px;
  border: 1px solid rgba(34, 197, 94, 0.35);
  border-radius: 999px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #86efac;
  background: rgba(34, 197, 94, 0.12);
}

.info-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: var(--space-6);
  margin-top: var(--space-6);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--border);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-weight: 600;
  color: var(--fg-2);
  font-size: var(--text-sm);
}

.info-value {
  color: var(--fg);
}

.mono {
  font-family: monospace;
  font-size: var(--text-sm);
}

@media (max-width: 768px) {
  .attendance-detail {
    padding: var(--space-6) var(--space-4);
  }

  .page-title {
    font-size: var(--text-xl);
  }

  .event-header,
  .participants-section,
  .info-section {
    padding: var(--space-6);
  }

  .event-meta {
    grid-template-columns: 1fr;
  }

  .action-bar {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    text-align: center;
  }
}
</style>
