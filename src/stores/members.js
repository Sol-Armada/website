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
                connectionStore.addListener('members', 'list').then((commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    resolve(commandResponse.result.members.map((m) => {
                        m.eventsAttended = commandResponse.result.event_counts[m.id]
                        return new Member(m)
                    }).sort((a, b) => {
                        if (a.rank.id < b.rank.id) {
                            if (a.rank.id == 0) {
                                return 1
                            }

                            return -1
                        }

                        if (a.rank.id > b.rank.id) {
                            if (b.rank.id == 0) {
                                return -1
                            }

                            return 1
                        }

                        return 0
                    }))
                }).catch((error) => {
                    console.log(error)
                })

                setTimeout(() => {
                    connectionStore.send('members', 'list', page)
                }, 500)
            })
        }
    }
})
