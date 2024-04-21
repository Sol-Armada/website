import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { Member } from "./classes"
import { useErrorStore } from "./error"

export const useMembersStore = defineStore("members", {
    actions: {
        async getMembers(page) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('members', 'list', (commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    resolve(commandResponse.result.map((m) => new Member(m)))
                })

                setTimeout(() => {
                    connectionStore.send('members', 'list', page)
                }, 1000)
            })
        }
    }
})
