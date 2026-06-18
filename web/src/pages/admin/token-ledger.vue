<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import type { TokenPeriodAnalytics } from '@/services/adminService'
  import { useTokenLedgerStore } from '@/stores/tokenLedger'

  const tokenLedgerStore = useTokenLedgerStore()
  const {
    loading,
    isRefreshing,
    error,
    transactions,
    ledgerSearch,
    page,
    pageInput,
    hasNextPage,
    analyticsLoading,
    analyticsRefreshing,
    analyticsError,
    analytics,
  } = storeToRefs(tokenLedgerStore)

  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null

  function formatTokenAmount(value: number): string {
    return `${value >= 0 ? '+' : '-'}${Math.abs(value)}`
  }

  function formatAverage(value: number): string {
    return value.toFixed(2)
  }

  function formatPeriodLabel(period: TokenPeriodAnalytics): string {
    const start = new Date(period.windowStart).toLocaleDateString()
    const end = new Date(period.windowEnd).toLocaleDateString()
    return `${start} - ${end}`
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
    void tokenLedgerStore.loadTokenLedger()
  })

  watch(ledgerSearch, () => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }

    searchDebounceTimer = setTimeout(() => {
      page.value = 1
      void tokenLedgerStore.loadTokenLedger({ background: true })
    }, 300)
  })

  onMounted(async() => {
    await tokenLedgerStore.initialize()
  })

  onBeforeUnmount(() => {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer)
    }
    tokenLedgerStore.dispose()
  })
</script>

<template>
  <PortalShell>
    <PageHeader
      subtitle=""
      title="Token Ledger"
    />

    <DataPanel description="Backend-aggregated statistics from the full token ledger." title="Ledger Analytics">
      <p
        v-if="analyticsRefreshing && !analyticsLoading"
        class="mb-3 text-xs font-medium uppercase tracking-wide text-on-surface-variant"
      >
        Refreshing analytics...
      </p>

      <StatePanel v-if="analyticsLoading" message="Loading ledger analytics..." title="Please wait" />

      <StatePanel v-else-if="analyticsError" :message="analyticsError" title="Analytics load failed" tone="error" />

      <template v-else-if="analytics">
        <h3 class="mb-3 text-sm font-semibold uppercase tracking-wide text-on-surface-variant">Week</h3>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <StatCard
            :detail="`Per member: ${formatAverage(analytics.week.averageEarningPerMember)} | per txn: ${formatAverage(analytics.week.averageEarningPerTransaction)}`"
            label="Avg Earning (Week)"
            :value="formatAverage(analytics.week.averageEarningPerTransaction)"
          />

          <StatCard
            :detail="`Per member: ${formatAverage(analytics.week.averageSpendingPerMember)} | per txn: ${formatAverage(analytics.week.averageSpendingPerTransaction)}`"
            label="Avg Spending (Week)"
            :value="formatAverage(analytics.week.averageSpendingPerTransaction)"
          />

          <StatCard
            :detail="formatPeriodLabel(analytics.week)"
            label="Week Earnings"
            :value="formatTokenAmount(analytics.week.totalEarnings)"
          />

          <StatCard
            :detail="formatPeriodLabel(analytics.week)"
            label="Week Spending"
            :value="formatTokenAmount(-analytics.week.totalSpending)"
          />
        </div>

        <h3 class="mb-3 mt-6 text-sm font-semibold uppercase tracking-wide text-on-surface-variant">Month</h3>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <StatCard
            :detail="`Per member: ${formatAverage(analytics.month.averageEarningPerMember)} | per txn: ${formatAverage(analytics.month.averageEarningPerTransaction)}`"
            label="Avg Earning (Month)"
            :value="formatAverage(analytics.month.averageEarningPerTransaction)"
          />

          <StatCard
            :detail="`Per member: ${formatAverage(analytics.month.averageSpendingPerMember)} | per txn: ${formatAverage(analytics.month.averageSpendingPerTransaction)}`"
            label="Avg Spending (Month)"
            :value="formatAverage(analytics.month.averageSpendingPerTransaction)"
          />

          <StatCard
            :detail="formatPeriodLabel(analytics.month)"
            label="Month Earnings"
            :value="formatTokenAmount(analytics.month.totalEarnings)"
          />

          <StatCard
            :detail="formatPeriodLabel(analytics.month)"
            label="Month Spending"
            :value="formatTokenAmount(-analytics.month.totalSpending)"
          />
        </div>

        <div class="mt-6 overflow-x-auto rounded-lg border border-subtle">
          <table class="w-full text-left text-sm text-on-surface">
            <thead class="bg-surface-variant/40 text-on-surface-variant">
              <tr>
                <th class="px-3 py-2">Reason</th>
                <th class="px-3 py-2">Transactions</th>
                <th class="px-3 py-2">Net Amount</th>
                <th class="px-3 py-2">Earnings</th>
                <th class="px-3 py-2">Spending</th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="reason in analytics.reasons" :key="reason.reason" class="border-t border-subtle">
                <td class="px-3 py-2">{{ reason.reason }}</td>
                <td class="px-3 py-2">{{ reason.transactionCount }}</td>

                <td class="px-3 py-2" :class="reason.netAmount >= 0 ? 'text-success' : 'text-error'">
                  {{ formatTokenAmount(reason.netAmount) }}
                </td>

                <td class="px-3 py-2 text-success">{{ formatTokenAmount(reason.totalEarnings) }}</td>
                <td class="px-3 py-2 text-error">{{ formatTokenAmount(-reason.totalSpending) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </DataPanel>

    <DataPanel description="Review credit and debit transactions across pages." title="Ledger Entries">
      <input
        v-model="ledgerSearch"
        class="w-full rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
        placeholder="Search ledger entries..."
        type="search"
      >

      <div class="mb-3 mt-2 h-0.5 w-full overflow-hidden rounded-full bg-surface-variant/40">
        <div
          class="h-full w-full bg-primary/80 transition-opacity duration-150"
          :class="isRefreshing && !loading ? 'animate-pulse opacity-100' : 'opacity-0'"
        />
      </div>

      <StatePanel v-if="loading" message="Loading token ledger..." title="Please wait" />

      <StatePanel v-else-if="error" :message="error" title="Ledger load failed" tone="error" />

      <div v-else-if="transactions.length > 0" class="overflow-x-auto rounded-lg border border-subtle">
        <table class="w-full text-left text-sm text-on-surface">
          <thead class="bg-surface-variant/40 text-on-surface-variant">
            <tr>
              <th class="px-3 py-2">Member</th>
              <th class="px-3 py-2">Amount</th>
              <th class="px-3 py-2">Reason</th>
              <th class="px-3 py-2">Event Name</th>
              <th class="px-3 py-2">Date</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="transaction in transactions" :key="transaction.id" class="border-t border-subtle">
              <td class="px-3 py-2">{{ transaction.memberName }}</td>

              <td class="px-3 py-2" :class="transaction.amount >= 0 ? 'text-success' : 'text-error'">
                {{ transaction.amount >= 0 ? '+' : '' }}{{ transaction.amount }}
              </td>

              <td class="px-3 py-2">{{ transaction.reason }}</td>

              <td class="px-3 py-2">{{ transaction.attendanceName || '—' }}</td>

              <td class="px-3 py-2">{{ new Date(transaction.createdAt).toLocaleDateString('en-US', {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                hour12: false
              }) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="text-sm text-on-surface-variant">
        {{ ledgerSearch ? 'No ledger entries matched your search.' : 'No ledger entries available.' }}
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
