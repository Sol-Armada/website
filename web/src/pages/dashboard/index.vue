<script setup lang="ts">
  import type { TokenTransaction } from '@/services/adminService'
  import { onMounted, ref, watch } from 'vue'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatCard from '@/components/ui/StatCard.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { type MemberDashboardData, memberService } from '@/services/memberService'

  const loading = ref(true)
  const ledgerLoading = ref(false)
  const error = ref<string | null>(null)
  const ledgerError = ref<string | null>(null)
  const dashboard = ref<MemberDashboardData | null>(null)
  const transactions = ref<TokenTransaction[]>([])
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)

  async function loadDashboard() {
    loading.value = true
    error.value = null

    try {
      dashboard.value = await memberService.getDashboard()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to load dashboard data'
    } finally {
      loading.value = false
    }
  }

  async function loadTokenLedger() {
    ledgerLoading.value = true
    ledgerError.value = null

    try {
      const response = await memberService.getTokenLedger(limit.value, page.value)
      transactions.value = response.records || []
      hasNextPage.value = transactions.value.length > limit.value
      ledgerError.value = null
    } catch(error_: any) {
      ledgerError.value = error_?.message || 'Failed to load token ledger'
      transactions.value = []
      hasNextPage.value = false
    } finally {
      ledgerLoading.value = false
    }
  }

  watch(page, () => {
    pageInput.value = String(page.value)
    void loadTokenLedger()
  })

  onMounted(async() => {
    await loadDashboard()
    await loadTokenLedger()
  })
</script>

<template>
  <PortalShell>
    <div class="w-full py-12 space-y-12">
      <PageHeader
        subtitle="Your personal token ledger, attendance, and activity feed"
        title="Member Dashboard"
      />

      <StatePanel v-if="loading" message="Loading dashboard..." title="Please wait" />

      <StatePanel v-else-if="error" :message="error" title="Dashboard load failed" tone="error" />

      <template v-else-if="dashboard">
        <!-- Stat Cards Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <StatCard detail="Operations attended" label="Event Attendance" :value="dashboard.attendance" />
          <StatCard detail="Total accumulated" label="Tokens Earned" :value="dashboard.tokens" />
          <StatCard label="Current Rank" :value="dashboard.rank" />
        </div>

        <!-- Token Ledger -->
        <section class="tactical-panel">
          <h2 class="section-title">Token Ledger</h2>

          <StatePanel v-if="ledgerLoading" message="Loading token ledger..." title="Please wait" />

          <StatePanel v-else-if="ledgerError" :message="ledgerError" title="Ledger load failed" tone="error" />

          <div v-else>
            <div class="overflow-x-auto rounded-lg border border-subtle">
              <table class="w-full text-left text-sm text-on-surface">
                <thead class="bg-surface-variant/40 text-on-surface-variant">
                  <tr>
                    <th class="px-3 py-2">Amount</th>
                    <th class="px-3 py-2">Reason</th>
                    <th class="px-3 py-2">Event Name</th>
                    <th class="px-3 py-2">Date</th>
                  </tr>
                </thead>

                <tbody>
                  <tr v-for="transaction in transactions" :key="transaction.id" class="border-t border-subtle">
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

            <p v-if="transactions.length === 0" class="mt-3 text-sm text-on-surface-variant">
              No token transactions yet.
            </p>

            <!-- Pagination Controls -->
            <div v-else-if="page > 1 || hasNextPage" class="mt-4 flex items-center justify-between gap-4">
              <div class="flex items-center gap-2">
                <button
                  class="px-3 py-1 text-sm border border-subtle rounded hover:bg-surface-variant disabled:opacity-50 disabled:cursor-not-allowed"
                  :disabled="page === 1 || ledgerLoading"
                  @click="page = Math.max(1, page - 1)"
                >
                  Previous
                </button>

                <input
                  v-model="pageInput"
                  class="w-16 px-2 py-1 text-sm text-center border border-subtle rounded bg-surface"
                  min="1"
                  type="number"
                  @change="page = Math.max(1, parseInt(pageInput) || 1)"
                >

                <span class="text-sm text-on-surface-variant">of ?</span>

                <button
                  class="px-3 py-1 text-sm border border-subtle rounded hover:bg-surface-variant disabled:opacity-50 disabled:cursor-not-allowed"
                  :disabled="!hasNextPage || ledgerLoading"
                  @click="page = page + 1"
                >
                  Next
                </button>
              </div>

              <div class="text-sm text-on-surface-variant">
                Page {{ page }} · {{ transactions.length }} results
              </div>
            </div>
          </div>
        </section>

        <!-- Recent Activity -->
        <!-- <section class="tactical-panel">
          <h2 class="section-title">Recent Activity</h2>

          <ul v-if="dashboard.recentActivity.length > 0" class="activity-list">
            <li v-for="activity in dashboard.recentActivity" :key="`${activity.type}-${activity.date}-${activity.title}`">
              <span class="activity-time">{{ new Date(activity.date).toLocaleDateString() }}</span>
              <span class="activity-desc">{{ activity.title }}</span>
            </li>
          </ul>

          <p v-else class="empty-state">No recent activity available.</p>
        </section> -->
      </template>
    </div>
  </PortalShell>
</template>

<style scoped>
.tactical-panel {
  position: relative;
  background: #121722;
  border: 1px solid #2a3447;
  border-radius: 16px;
  padding: 32px;
  overflow: hidden;
}

.tactical-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #E6A82D 0%, #0041FF 50%, transparent 100%);
  opacity: 0.8;
  border-radius: 16px 16px 0 0;
}

.section-title {
  font-size: 24px;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 24px;
}

.token-ledger-table {
  overflow-x: auto;
}

.token-ledger-table table {
  width: 100%;
  border-collapse: collapse;
}

.token-ledger-table thead tr {
  border-bottom: 2px solid;
  border-image: linear-gradient(90deg, #E6A82D 0%, #0041FF 50%, transparent 100%) 1;
}

.token-ledger-table th {
  padding: 12px 16px;
  text-align: left;
  font-family: var(--font-mono);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-2);
  font-weight: 600;
}

.token-ledger-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-soft);
  font-size: 14px;
}

.token-ledger-table tbody tr:hover {
  background: rgba(230, 168, 45, 0.05);
}

.token-ledger-table .date,
.token-ledger-table .balance {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}

.token-ledger-table .amount-positive {
  color: var(--success);
  font-family: var(--font-mono);
  font-weight: 600;
}

.token-ledger-table .amount-negative {
  color: var(--danger);
  font-family: var(--font-mono);
  font-weight: 600;
}

.token-ledger-table .type-badge {
  display: inline-block;
  padding: 4px 8px;
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.3);
  border-radius: 4px;
  font-size: 11px;
  font-family: var(--font-mono);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #60a5fa;
}

.activity-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.activity-list li {
  display: flex;
  gap: 12px;
  align-items: baseline;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-soft);
}

.activity-list li:last-child {
  border-bottom: none;
}

.activity-time {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--muted);
  min-width: 100px;
}

.activity-desc {
  font-size: 14px;
  color: var(--fg-2);
}

.empty-state {
  text-align: center;
  padding: 24px;
  color: var(--muted);
  font-size: 14px;
}
</style>
