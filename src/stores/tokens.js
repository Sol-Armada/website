import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { useErrorStore } from "./error"
import { useMembersStore } from "./members"
import { Token } from "./classes"

export const useTokensStore = defineStore("tokens", {
    state: () => ({
        tokens: []
    }),
    actions: {
        async getTokenRecords(page) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                const membersStore = useMembersStore()
                connectionStore.addListener('tokens', 'list').then(async (commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    const tokenPromises = commandResponse.result.map(async (tkn) => {
                        try {
                            const member = await membersStore.getMember(tkn.member_id)
                            tkn.member = member.name

                            if (tkn.giver_id != null) {
                                const giver = await membersStore.getMember(tkn.giver_id)
                                tkn.giver = giver.name
                            }

                            tkn = new Token(tkn)
                        } catch (error) {
                            console.log(error)
                        }
                    })

                    await Promise.all(tokenPromises)

                    this.tokens = commandResponse.result
                    resolve(true)
                }).catch((error) => {
                    console.log(error)
                    errorStore.$patch({ error: error.message, show: true })
                })

                setTimeout(() => {
                    connectionStore.send('tokens', 'list', page)
                }, 500)
            })
        }
    }
})
