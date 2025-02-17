import { defineStore, acceptHMRUpdate } from 'pinia'
import { ref } from 'vue'
import { CommandResponse } from "./classes"
import { useErrorStore } from "./error"
import { useAppStore } from "./app"

// const URL = process.env.NODE_ENV === "production" ? import.meta.env.VITE_WEBSOCKET_URL : "ws://localhost:3001/ws"
const URL = import.meta.env.VITE_WEBSOCKET_URL

export const useConnectionStore = defineStore("connection", () => {
    const isConnected = ref(false)
    const socket = ref(null)
    const appStore = useAppStore()

    function bindEvents() {
        if (isConnected.value) {
            console.log("already connected")
            return
        }

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

            console.log("connected")
        }

        socket.value.onclose = () => {
            console.log("reconnecting...")

            isConnected.value = false

            if (!errorStore.show && errorStore.reason != "websocket") {
                errorStore.$patch({ error: "Lost connection to the backend. Trying to reconnect...", show: true, loading: true, timeout: -1, closeable: false, reason: "websocket" })
            }

            setTimeout(() => {
                this.connect()
            }, 5000)
        }
    }

    async function send(command, action, arg) {
        while (isConnected.value == false) {
            await new Promise(resolve => setTimeout(resolve, 1000))
        }

        socket.value.send(command + "|" + action + ":" + arg + ":" + appStore.token)
    }

    async function addListener(thing, action) {
        return new Promise((resolve, reject) => {
            socket.value.addEventListener("message", (event) => {
                const commandResponse = new CommandResponse(event.data)

                if (commandResponse.error) {
                    reject(commandResponse.error)
                    return
                }

                if (commandResponse.thing === thing && commandResponse.action === action) {
                    resolve(commandResponse)
                }
            })
        })
    }

    function addForeverListener(thing, action, callback) {
        socket.value.addEventListener("message", (event) => {
            const commandResponse = new CommandResponse(event.data)

            if (commandResponse.thing === thing && commandResponse.action === action) {
                callback(commandResponse)
            }
        })
    }

    return {
        bindEvents,
        connect,
        send,
        addListener,
        addForeverListener,
        isConnected
    }
})

if (import.meta.hot) {
    import.meta.hot.accept(acceptHMRUpdate(useConnectionStore, import.meta.hot))
}
