import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { Member } from "./classes"
import { useErrorStore } from "./error"

export const useMembersStore = defineStore("members", {
    state: () => ({
        members: [],
    }),
    actions: {
        bindEvents() { this.getMembers() },
        getMembers() {
            const errorStore = useErrorStore()
            const connectionStore = useConnectionStore()

            connectionStore.addListener('members', 'list', (commandResponse) => {
                // handle errors
                if (commandResponse.error) {
                    errorStore.$patch({ error: commandResponse.error, show: true })
                    return
                }

                this.$patch({ members: commandResponse.result.map((m) => new Member(m)) })
            })

            connectionStore.send('members', 'list', 1)
        }
    }
})
