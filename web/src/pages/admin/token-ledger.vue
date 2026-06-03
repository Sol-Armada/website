<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import DataPanel from '@/components/ui/DataPanel.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import {
    adminService,
    type TokenLedgerAnalytics,
    type TokenPeriodAnalytics,
    type TokenTransaction,
  } from '@/services/adminService'
  import { WS_TOPIC_ADMIN_TOKEN_LEDGER, wsClient } from '@/services/wsClient'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const transactions = ref<TokenTransaction[]>([])
  const page = ref(1)
  const limit = ref(25)
  const hasNextPage = ref(false)

  const analyticsLoading = ref(true)
  const analyticsError = ref<string | null>(null)
  const analytics = ref<TokenLedgerAnalytics | null>(null)
  let refreshTimer: number | null = null
  const unsubscribers: Array<() => void> = []

  function scheduleRefresh() {
    if (refreshTimer !== null) {
      window.clearTimeout(refreshTimer)
    }
    refreshTimer = window.setTimeout(() => {
      refreshTimer = null
      void Promise.all([loadTokenLedger(), loadAnalytics()])
    }, 400)
  }

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

  async function loadAnalytics(): Promise<void> {
    analyticsLoading.value = true
    analyticsError.value = null

    try {
      analytics.value = await adminService.getTokenLedgerAnalytics()
    } catch(error_: any) {
      analyticsError.value = error_?.message || 'Failed to load analytics'
      analytics.value = null
    } finally {
      analyticsLoading.value = false
    }
  }

  async function loadTokenLedger(): Promise<void> {
    loading.value = true
    error.value = null

    try {
      const response = await adminService.getTokenLedger(limit.value, page.value)
      transactions.value = response.records || []
      hasNextPage.value = transactions.value.length === limit.value
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to load token ledger'
      hasNextPage.value = false
    } finally {
      loading.value = false
    }
  }

  function goToPreviousPage(): void {
    if (page.value <= 1 || loading.value) return

    page.value -= 1
  }

  function goToNextPage(): void {
    if (!hasNextPage.value || loading.value) return

    page.value += 1
  }

  watch(page, () => {
    void loadTokenLedger()
  })

  onMounted(async() => {
    await Promise.all([loadTokenLedger(), loadAnalytics()])
    unsubscribers.push(wsClient.onTopic(WS_TOPIC_ADMIN_TOKEN_LEDGER, scheduleRefresh))
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
      subtitle="Token activity analytics with weekly and monthly earning and spending patterns."
      title="Token Ledger"
    />

    <DataPanel
      description="Backend-aggregated statistics from the full token ledger."
      title="Ledger Analytics"
    >
      <StatePanel v-if="analyticsLoading" message="Loading ledger analytics..." title="Please wait" />

      <StatePanel
        v-else-if="analyticsError"
        :message="analyticsError"
        title="Analytics load failed"
        tone="error"
      />

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
              <tr
                v-for="reason in analytics.reasons"
                :key="reason.reason"
                class="border-t border-subtle"
              >
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

    <DataPanel
      description="Review credit and debit transactions across pages."
      title="Ledger Entries"
    >
      <StatePanel v-if="loading" message="Loading token ledger..." title="Please wait" />

      <StatePanel
        v-else-if="error"
        :message="error"
        title="Ledger load failed"
        tone="error"
      />

      <div v-else-if="transactions.length > 0" class="overflow-x-auto rounded-lg border border-subtle">
        <table class="w-full text-left text-sm text-on-surface">
          <thead class="bg-surface-variant/40 text-on-surface-variant">
            <tr>
              <th class="px-3 py-2">Member</th>
              <th class="px-3 py-2">Amount</th>
              <th class="px-3 py-2">Reason</th>
              <th class="px-3 py-2">Date</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="transaction in transactions" :key="transaction.id" class="border-t border-subtle">
              <td class="px-3 py-2">{{ transaction.memberId }}</td>

              <td class="px-3 py-2" :class="transaction.amount >= 0 ? 'text-success' : 'text-error'">
                {{ transaction.amount >= 0 ? '+' : '' }}{{ transaction.amount }}
              </td>

              <td class="px-3 py-2">{{ transaction.reason }}</td>
              <td class="px-3 py-2">{{ new Date(transaction.createdAt).toLocaleDateString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="text-sm text-on-surface-variant">No ledger entries available.</p>

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
