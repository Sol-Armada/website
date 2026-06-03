// Utilities
import { defineStore } from 'pinia'

export type RealtimeConnectionState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected'

export const useAppStore = defineStore('app', {
  state: () => ({
    realtimeState: 'idle' as RealtimeConnectionState,
    realtimeLastError: '' as string,
    realtimeConnectedAt: null as string | null,
  }),
  actions: {
    setRealtimeState(state: RealtimeConnectionState) {
      this.realtimeState = state
      if (state === 'connected') {
        this.realtimeConnectedAt = new Date().toISOString()
        this.realtimeLastError = ''
      }
    },
    setRealtimeError(message: string) {
      this.realtimeLastError = message
    },
  },
})
