<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref } from 'vue'
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
      loadOverview()
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
      subtitle="Skeleton admin KPI surface for system health and operational totals."
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
        <DataPanel description="Unique participants in tracked events." title="Unique Attendees">
          <p class="text-2xl font-semibold text-on-surface">{{ overview.uniqueAttendees }}</p>
        </DataPanel>

        <DataPanel description="Average attendees per tracked event." title="Average Attendance">
          <p class="text-2xl font-semibold text-on-surface">{{ overview.averageAttendance }}</p>
        </DataPanel>
      </section>
    </template>
  </PortalShell>
</template>
