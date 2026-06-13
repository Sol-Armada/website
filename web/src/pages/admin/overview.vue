<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref } from 'vue'
  import { RouterLink } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { type AdminOverviewData, adminService } from '@/services/adminService'
  import { WS_TOPIC_ADMIN_ATTENDANCE, WS_TOPIC_ADMIN_MEMBERS, WS_TOPIC_ADMIN_TOKEN_LEDGER, wsClient } from '@/services/wsClient'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const overview = ref<AdminOverviewData | null>(null)
  let refreshTimer: number | null = null
  const unsubscribers: Array<() => void> = []

  async function loadOverview() {
    loading.value = true
    error.value = null

    try {
      overview.value = await adminService.getOverview()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to load admin overview'
    } finally {
      loading.value = false
    }
  }

  function scheduleRefresh() {
    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
    }
    refreshTimer = window.setTimeout(() => {
      refreshTimer = null
      // loadOverview()
    }, 400)
  }

  onMounted(async() => {
    await loadOverview()

    unsubscribers.push(
      wsClient.onTopic(WS_TOPIC_ADMIN_MEMBERS, scheduleRefresh),
      wsClient.onTopic(WS_TOPIC_ADMIN_ATTENDANCE, scheduleRefresh),
      wsClient.onTopic(WS_TOPIC_ADMIN_TOKEN_LEDGER, scheduleRefresh),
    )
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
</script>

<template>
  <PortalShell>
    <PageHeader
      subtitle="Organization metrics, admin tools, and operational totals."
      title="Admin Overview"
    />

    <StatePanel v-if="loading" message="Loading overview..." title="Please wait" />

    <StatePanel
      v-else-if="error"
      :message="error"
      title="Overview load failed"
      tone="error"
    />

    <template v-else-if="overview">
      <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard detail="Current total" label="Total Members" :value="overview.totalMembers" />
        <StatCard detail="Tracked events" label="Events" :value="overview.totalEvents" />
        <StatCard detail="Total issued" label="Tokens" :value="overview.totalTokens" />
        <StatCard detail="Active this month" label="Active Members" :value="overview.activeThisMonth" />
      </section>

      <section class="mt-6 grid gap-4 lg:grid-cols-2">
        <StatCard detail="Unique participants in tracked events" label="Unique Attendees" :value="overview.uniqueAttendees" />
        <StatCard detail="Average attendees per tracked event" label="Average Attendance" :value="overview.averageAttendance" />
      </section>

      <!-- <section class="tactical-panel mt-6 p-5">
        <h2 class="text-lg font-semibold text-on-surface">Quick Links</h2>
        <p class="mt-2 text-sm text-on-surface-variant">Administrative tools and management functions</p>

        <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <RouterLink
            class="flex items-center gap-3 rounded-lg border border-divider bg-surface/50 px-4 py-3 text-sm font-medium text-on-surface hover:border-primary hover:bg-surface transition-all"
            to="/admin/members"
          >
            <span>👥</span>
            <span>Manage Members</span>
          </RouterLink>

          <RouterLink
            class="flex items-center gap-3 rounded-lg border border-divider bg-surface/50 px-4 py-3 text-sm font-medium text-on-surface hover:border-primary hover:bg-surface transition-all"
            to="/admin/attendance"
          >
            <span>📅</span>
            <span>Attendance Reports</span>
          </RouterLink>

          <RouterLink
            class="flex items-center gap-3 rounded-lg border border-divider bg-surface/50 px-4 py-3 text-sm font-medium text-on-surface hover:border-primary hover:bg-surface transition-all"
            to="/admin/token-ledger"
          >
            <span>🪙</span>
            <span>Token Ledger</span>
          </RouterLink>

          <button
            class="flex items-center gap-3 rounded-lg border border-divider bg-surface/50 px-4 py-3 text-sm font-medium text-on-surface hover:border-primary hover:bg-surface transition-all"
            disabled
          >
            <span>👤</span>
            <span>Role Management</span>
          </button>

          <button
            class="flex items-center gap-3 rounded-lg border border-divider bg-surface/50 px-4 py-3 text-sm font-medium text-on-surface hover:border-primary hover:bg-surface transition-all"
            disabled
          >
            <span>⚙️</span>
            <span>Organization Settings</span>
          </button>
        </div>
      </section> -->
    </template>
  </PortalShell>
</template>
