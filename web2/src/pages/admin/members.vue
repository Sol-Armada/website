<script setup lang="ts">
  import { onMounted, onUnmounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService, type MemberSummary } from '@/services/adminService'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const members = ref<MemberSummary[]>([])
  const search = ref('')
  const page = ref(1)
  const limit = ref(25)
  const hasNextPage = ref(false)

  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null

  async function loadMembers (): Promise<void> {
    loading.value = true
    error.value = null

    try {
      const response = await adminService.getMembers(limit.value, page.value, search.value || undefined)
      members.value = response.members || []
      hasNextPage.value = members.value.length === limit.value
    } catch (error_: any) {
      error.value = error_?.message || 'Failed to load members'
      hasNextPage.value = false
    } finally {
      loading.value = false
    }
  }

  function goToPreviousPage (): void {
    if (page.value <= 1 || loading.value) return

    page.value -= 1
  }

  function goToNextPage (): void {
    if (!hasNextPage.value || loading.value) return

    page.value += 1
  }

  watch(page, () => {
    void loadMembers()
  })

  watch(search, () => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)

    searchDebounceTimer = setTimeout(() => {
      page.value = 1
      void loadMembers()
    }, 300)
  })

  onMounted(async () => {
    await loadMembers()
  })

  onUnmounted(() => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })
</script>

<template>
  <PortalShell>
    <PageHeader
      subtitle="Searchable member directory with rank tags and simple paging."
      title="Members"
    />

    <DataPanel
      description="Browse members with search and paging controls."
      title="Member Directory"
    >
      <input
        v-model="search"
        class="mb-3 w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
        placeholder="Search members..."
        type="search"
      >

      <StatePanel v-if="loading" message="Loading members..." title="Please wait" />

      <StatePanel
        v-else-if="error"
        :message="error"
        title="Members load failed"
        tone="error"
      />

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
            @click="goToPreviousPage"
          >
            Previous
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
