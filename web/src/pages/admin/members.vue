<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { onBeforeUnmount, onMounted, onUnmounted, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { useMembersStore } from '@/stores/members'

  const membersStore = useMembersStore()
  const {
    loading,
    isRefreshing,
    error,
    members,
    search,
    page,
    pageInput,
    hasNextPage,
  } = storeToRefs(membersStore)

  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let nextPageLoadMode: 'background' | 'blocking' = 'blocking'

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
    const shouldUseBackgroundLoad = nextPageLoadMode === 'background'
    nextPageLoadMode = 'blocking'
    void membersStore.loadMembers({ background: shouldUseBackgroundLoad })
  })

  watch(search, () => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)

    searchDebounceTimer = setTimeout(() => {
      nextPageLoadMode = 'background'

      if (page.value !== 1) {
        page.value = 1
        return
      }

      void membersStore.loadMembers({ background: true })
    }, 300)
  })

  onMounted(async() => {
    await membersStore.initialize()
  })

  onBeforeUnmount(() => {
    membersStore.dispose()
  })

  onUnmounted(() => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })
</script>

<template>
  <PortalShell>
    <PageHeader subtitle="" title="Members" />

    <DataPanel description="Review member stats, attendance count, and current token balances." title="Member Directory">
      <input
        v-model="search"
        class="w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
        placeholder="Search members..."
        type="search"
      >

      <div class="mb-3 mt-2 h-0.5 w-full overflow-hidden rounded-full bg-surface-variant/40">
        <div
          class="h-full w-full bg-primary/80 transition-opacity duration-150"
          :class="isRefreshing && !loading ? 'animate-pulse opacity-100' : 'opacity-0'"
        />
      </div>

      <StatePanel v-if="error" :message="error" title="Members load failed" tone="error" />

      <div v-else-if="members.length > 0" class="overflow-x-auto rounded-lg border border-subtle mt-2">
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

      <p v-else class="text-sm text-on-surface-variant">
        {{ search ? 'No members matched your search.' : 'No members found.' }}
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
