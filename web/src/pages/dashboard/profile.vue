<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { type MemberProfileData, memberService } from '@/services/memberService'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const profile = ref<MemberProfileData | null>(null)

  onMounted(async() => {
    loading.value = true
    error.value = null

    try {
      profile.value = await memberService.getProfile()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to load profile data'
    } finally {
      loading.value = false
    }
  })
</script>

<template>
  <PortalShell>
    <PageHeader
      subtitle="Skeleton profile view. Identity and metrics are currently mock-backed."
      title="My Profile"
    />

    <StatePanel v-if="loading" message="Loading profile..." title="Please wait" />

    <StatePanel
      v-else-if="error"
      :message="error"
      title="Profile load failed"
      tone="error"
    />

    <section v-else-if="profile" class="grid gap-4 lg:grid-cols-3">
      <DataPanel description="Current authenticated member profile." title="Identity">
        <p class="text-base font-semibold text-on-surface">{{ profile.username }}</p>
        <p class="mt-2 text-sm text-on-surface-variant">Rank: {{ profile.rank }}</p>
        <p v-if="profile.rsiHandle" class="mt-2 text-sm text-on-surface-variant">RSI: {{ profile.rsiHandle }}</p>
        <p class="mt-2 text-sm text-on-surface-variant">Validated: {{ profile.validated ? 'Yes' : 'No' }}</p>
      </DataPanel>

      <StatCard detail="Events attended" label="Attendance" :value="profile.attendanceCount" />
      <StatCard detail="Current token balance" label="Token Balance" :value="profile.tokensBalance" />
    </section>
  </PortalShell>
</template>
