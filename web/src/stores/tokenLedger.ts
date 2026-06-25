import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminService, type TokenLedgerAnalytics, type TokenTransaction } from '@/services/adminService'
import { WS_TOPIC_ADMIN_TOKEN_LEDGER, wsClient } from '@/services/wsClient'
import { createRefreshScheduler } from '@/stores/refreshScheduler'
import { createRequestQueue } from '@/stores/requestQueue'

export const useTokenLedgerStore = defineStore('token-ledger', () => {
  const loading = ref(true)
  const isRefreshing = ref(false)
  const error = ref<string | null>(null)
  const transactions = ref<TokenTransaction[]>([])
  const ledgerSearch = ref('')
  const page = ref(1)
  const pageInput = ref('1')
  const limit = ref(25)
  const hasNextPage = ref(false)

  const analyticsLoading = ref(true)
  const analyticsRefreshing = ref(false)
  const analyticsError = ref<string | null>(null)
  const analytics = ref<TokenLedgerAnalytics | null>(null)

  const refreshScheduler = createRefreshScheduler()
  const ledgerRequestQueue = createRequestQueue()
  const analyticsRequestQueue = createRequestQueue()
  let unsubscribeTokenLedger: (() => void) | null = null

  function scheduleRefresh() {
    refreshScheduler.schedule(() => {
      if (ledgerSearch.value.trim() !== '') {
        void loadAnalytics({ background: true })
        return
      }

      void Promise.all([
        loadTokenLedger({ background: true }),
        loadAnalytics({ background: true }),
      ])
    })
  }

  async function loadAnalytics(options: { background?: boolean } = {}): Promise<void> {
    await analyticsRequestQueue.run(options, async isBackground => {
      if (isBackground) {
        analyticsRefreshing.value = true
      } else {
        analyticsLoading.value = true
        analyticsError.value = null
      }

      try {
        analytics.value = await adminService.getTokenLedgerAnalytics()
        analyticsError.value = null
      } catch(error_: any) {
        if (!isBackground || analytics.value === null) {
          analyticsError.value = error_?.message || 'Failed to load analytics'
          analytics.value = null
        }
      } finally {
        if (isBackground) {
          analyticsRefreshing.value = false
        } else {
          analyticsLoading.value = false
        }
      }
    })
  }

  async function loadTokenLedger(options: { background?: boolean } = {}): Promise<void> {
    await ledgerRequestQueue.run(options, async isBackground => {
      if (isBackground) {
        isRefreshing.value = true
      } else {
        loading.value = true
        error.value = null
      }

      try {
        const response = await adminService.getTokenLedger(limit.value, page.value, ledgerSearch.value || undefined)
        transactions.value = response.records || []
        hasNextPage.value = transactions.value.length === limit.value
        error.value = null
      } catch(error_: any) {
        if (!isBackground || transactions.value.length === 0) {
          error.value = error_?.message || 'Failed to load token ledger'
          hasNextPage.value = false
        }
      } finally {
        if (isBackground) {
          isRefreshing.value = false
        } else {
          loading.value = false
        }
      }
    })
  }

  async function initialize(): Promise<void> {
    if (unsubscribeTokenLedger === null) {
      unsubscribeTokenLedger = wsClient.onTopic(WS_TOPIC_ADMIN_TOKEN_LEDGER, scheduleRefresh)
    }

    await Promise.all([loadTokenLedger(), loadAnalytics()])
  }

  function dispose(): void {
    refreshScheduler.clear()

    ledgerRequestQueue.clear()
    analyticsRequestQueue.clear()

    if (unsubscribeTokenLedger !== null) {
      unsubscribeTokenLedger()
      unsubscribeTokenLedger = null
    }
  }

  return {
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
    loadAnalytics,
    loadTokenLedger,
    initialize,
    dispose,
  }
})
