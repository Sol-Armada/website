<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { type MemberDashboardData, memberService } from '@/services/memberService'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const dashboard = ref<MemberDashboardData | null>(null)

  onMounted(async () => {
    loading.value = true
    error.value = null

    try {
      dashboard.value = await memberService.getDashboard()
    } catch (error_: any) {
      error.value = error_?.message || 'Failed to load dashboard data'
    } finally {
      loading.value = false
    }
  })
</script>

<template>
  <PortalShell>
    <PageHeader
      subtitle="Skeleton surface for attendance, tokens, rank, and activity data."
      title="Member Dashboard"
    />

    <StatePanel v-if="loading" message="Loading dashboard..." title="Please wait" />

    <StatePanel
      v-else-if="error"
      :message="error"
      title="Dashboard load failed"
      tone="error"
    />

    <template v-else-if="dashboard">
      <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard detail="Events attended" label="Attendance" :value="dashboard.attendance" />
        <StatCard detail="Current token balance" label="Tokens" :value="dashboard.tokens" />
        <StatCard detail="Current member rank" label="Rank" :value="dashboard.rank" />
      </section>

      <section class="mt-6 grid gap-4 lg:grid-cols-2">
        <DataPanel description="Latest member timeline entries from backend." title="Recent Activity">
          <ul v-if="dashboard.recentActivity.length > 0" class="space-y-2 text-sm text-on-surface-variant">
            <li
              v-for="activity in dashboard.recentActivity"
              :key="`${activity.type}-${activity.date}-${activity.title}`"
            >
              <span class="font-semibold text-on-surface">{{ activity.title }}</span>
              <span class="ml-2">{{ new Date(activity.date).toLocaleDateString() }}</span>
            </li>
          </ul>

          <p v-else class="text-sm text-on-surface-variant">No recent activity available.</p>
        </DataPanel>

        <StatePanel
          message="Member dashboard is now API-backed. Additional analytics tiles can be added here next."
          title="Integration status"
          tone="success"
        />
      </section>
    </template>
  </PortalShell>
</template>
