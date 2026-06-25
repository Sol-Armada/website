import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type AdminOverviewData, adminService } from '@/services/adminService'
import { WS_TOPIC_ADMIN_ATTENDANCE, WS_TOPIC_ADMIN_MEMBERS, WS_TOPIC_ADMIN_TOKEN_LEDGER, wsClient } from '@/services/wsClient'
import { createRefreshScheduler } from '@/stores/refreshScheduler'

export const useOverviewStore = defineStore('overview', () => {
    const loading = ref(true)
    const error = ref<string | null>(null)
    const overview = ref<AdminOverviewData | null>(null)
    const refreshScheduler = createRefreshScheduler()
    const unsubscribers: Array<() => void> = []

    async function loadOverview(): Promise<void> {
        loading.value = true
        error.value = null

        try {
            overview.value = await adminService.getOverview()
        } catch (error_: any) {
            error.value = error_?.message || 'Failed to load admin overview'
        } finally {
            loading.value = false
        }
    }

    function scheduleRefresh(): void {
        refreshScheduler.schedule(() => {
            void loadOverview()
        })
    }

    async function initialize(): Promise<void> {
        if (unsubscribers.length === 0) {
            unsubscribers.push(
                wsClient.onTopic(WS_TOPIC_ADMIN_MEMBERS, scheduleRefresh),
                wsClient.onTopic(WS_TOPIC_ADMIN_ATTENDANCE, scheduleRefresh),
                wsClient.onTopic(WS_TOPIC_ADMIN_TOKEN_LEDGER, scheduleRefresh),
            )
        }

        await loadOverview()
    }

    function dispose(): void {
        refreshScheduler.clear()

        for (const unsubscribe of unsubscribers) {
            unsubscribe()
        }
        unsubscribers.splice(0, unsubscribers.length)
    }

    return {
        loading,
        error,
        overview,
        loadOverview,
        initialize,
        dispose,
    }
})
