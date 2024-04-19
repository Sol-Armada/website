import { defineStore, acceptHMRUpdate } from 'pinia'
import { CommandResponse } from "./classes"
import { useErrorStore } from "./error"

// const URL = process.env.NODE_ENV === "production" ? import.meta.env.VITE_WEBSOCKET_URL : "ws://localhost:3001/ws"
const URL = import.meta.env.VITE_WEBSOCKET_URL

export const useConnectionStore = defineStore("connection", () => {
    const isConnected = ref(false)
    const socket = ref(null)

    function bindEvents() {
        this.connect()
    }

    function connect() {
        const errorStore = useErrorStore()

        socket.value = new WebSocket(URL)

        socket.value.onopen = () => {
            isConnected.value = true

            if (errorStore.show && errorStore.reason == "websocket") {
                errorStore.reset()
            }

            console.log("Connected")
        }

        socket.value.onclose = () => {
            console.log("Reconnecting...")

            isConnected.value = false

            if (!errorStore.show && errorStore.reason != "websocket") {
                errorStore.$patch({ error: "Lost connection to the backend. Trying to reconnect...", show: true, loading: true, timeout: -1, closeable: false, reason: "websocket" })
            }

            setTimeout(() => {
                this.connect()
            }, 5000)
        }
    }

    function send(command) {
        let sent = false
        for (let i = 0; i < 5; i++) {
            setTimeout(() => {
                if (isConnected.value && !sent) {
                    socket.value.send(command)
                    sent = true
                }
            }, 1000)
            if (sent) {
                break
            }
        }
    }

    function addListener(thing, callback) {
        socket.value.addEventListener("message", (event) => {
            const commandResponse = new CommandResponse(event.data)

            if (commandResponse.thing === thing) {
                callback(commandResponse)
            }
        })
    }

    return {
        bindEvents,
        connect,
        send,
        addListener,
        isConnected
    }
})

if (import.meta.hot) {
 import.meta.hot.accept(acceptHMRUpdate(useConnectionStore, import.meta.hot))
}
