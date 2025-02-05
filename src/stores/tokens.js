import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { Token } from "./classes"
import { useErrorStore } from "./error"

export const useTokensStore = defineStore("tokens", {
    state: () => ({
        tokens: new Map()
    }),
    actions: {
        async getTokenRecords(page) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('tokens', 'list').then((commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    this.tokens = new Map(Object.entries(commandResponse.result))
                    for (const [key, value] of this.tokens) {
                        value.forEach((token, i) => {
                            value[i] = new Token(token)
                        })
                        this.tokens.set(key, value)
                    }
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
