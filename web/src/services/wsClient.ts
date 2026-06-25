import { useAppStore } from '@/stores/app'

export const WS_TOPIC_SYSTEM_HEALTH = 'system.health'
export const WS_TOPIC_ADMIN_MEMBERS = 'admin.members.updated'
export const WS_TOPIC_ADMIN_ATTENDANCE = 'admin.attendance.updated'
export const WS_TOPIC_ADMIN_TOKEN_LEDGER = 'admin.token_ledger.updated'

const WS_BASE_URL = trimTrailingSlash(import.meta.env.VITE_WS_BASE_URL || '')

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export interface RealtimeEnvelope {
  type: string
  topic: string
  sequence?: number
  timestamp: string
  payload?: any
}

type TopicHandler = (event: RealtimeEnvelope) => void

class WebSocketClient {
  private socket: WebSocket | null = null
  private shouldReconnect = false
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private requestedTopics: string[] = [WS_TOPIC_SYSTEM_HEALTH]
  private handlers = new Map<string, Set<TopicHandler>>()

  connect(topics: string[] = [WS_TOPIC_SYSTEM_HEALTH]) {
    this.shouldReconnect = true
    this.requestedTopics = uniqueTopics([WS_TOPIC_SYSTEM_HEALTH, ...topics])

    if (this.socket?.readyState === WebSocket.OPEN) {
      this.sendSubscribe()
      return
    }

    if (this.socket?.readyState === WebSocket.CONNECTING) {
      return
    }

    this.openSocket()
  }

  disconnect() {
    this.shouldReconnect = false
    this.clearReconnectTimer()

    if (this.socket) {
      this.socket.close(1000, 'client disconnect')
      this.socket = null
    }

    const appStore = useAppStore()
    appStore.setRealtimeState('idle')
  }

  onTopic(topic: string, handler: TopicHandler): () => void {
    const topicHandlers = this.handlers.get(topic) ?? new Set<TopicHandler>()
    topicHandlers.add(handler)
    this.handlers.set(topic, topicHandlers)

    return () => {
      const handlers = this.handlers.get(topic)
      if (!handlers) {
        return
      }
      handlers.delete(handler)
      if (handlers.size === 0) {
        this.handlers.delete(topic)
      }
    }
  }

  private openSocket() {
    const appStore = useAppStore()
    appStore.setRealtimeState(this.reconnectAttempts > 0 ? 'reconnecting' : 'connecting')

    const wsUrl = this.buildWebSocketUrl()
    this.socket = new WebSocket(wsUrl)

    this.socket.addEventListener('open', () => {
      const store = useAppStore()
      this.reconnectAttempts = 0
      store.setRealtimeState('connected')
      this.sendSubscribe()
    })

    this.socket.addEventListener('message', event => {
      try {
        const message = JSON.parse(event.data) as RealtimeEnvelope
        this.dispatch(message)
      } catch {
        // Ignore malformed messages.
      }
    })

    this.socket.addEventListener('error', () => {
      const store = useAppStore()
      store.setRealtimeError('Realtime connection error')
    })

    this.socket.addEventListener('close', () => {
      this.socket = null

      if (!this.shouldReconnect) {
        return
      }

      const store = useAppStore()
      store.setRealtimeState('disconnected')
      this.scheduleReconnect()
    })
  }

  private scheduleReconnect() {
    this.clearReconnectTimer()
    this.reconnectAttempts += 1

    const baseDelay = Math.min(30_000, 1000 * (2 ** Math.min(this.reconnectAttempts, 6)))
    const jitter = Math.floor(Math.random() * 300)
    const delayMs = baseDelay + jitter

    this.reconnectTimer = window.setTimeout(() => {
      this.openSocket()
    }, delayMs)
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer === null) {
      return
    }
    window.clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
  }

  private sendSubscribe() {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return
    }

    this.socket.send(JSON.stringify({
      type: 'subscribe',
      topics: this.requestedTopics,
    }))
  }

  private dispatch(message: RealtimeEnvelope) {
    if (import.meta.env.DEV) {
      const operation = message.payload?.operation
      console.debug('[ws:event]', {
        type: message.type,
        topic: message.topic,
        operation,
        sequence: message.sequence,
        payload: message.payload,
      })
    }

    const handlers = this.handlers.get(message.topic)
    if (!handlers) {
      return
    }

    for (const handler of handlers) {
      handler(message)
    }
  }

  private buildWebSocketUrl(): string {
    if (WS_BASE_URL) {
      return `${WS_BASE_URL}/api/ws`
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/api/ws`
  }
}

function uniqueTopics(topics: string[]): string[] {
  return [...new Set(topics)]
}

export const wsClient = new WebSocketClient()
