// Utilities
import { defineStore } from 'pinia'
import { useConnectionStore } from './connection'
import { useErrorStore } from './error'
import { Member } from './classes'

export const useAppStore = defineStore('app', {
    state: () => ({
        loggedIn: null,
        /** @type {string} */
        authCode: null,
        /** @type {string} */
        accessToken: null,
        /** @type {Member} */
        me: null
    }),
    actions: {
        bindEvents() {
            const daStorage = localStorage.getItem("discord_access")
            const meStorage = localStorage.getItem("me")

            if (meStorage) {
                this.me = new Member(JSON.parse(meStorage))
            }

            if (daStorage == "null") {
                localStorage.removeItem("discord_access")
                return
            }

            const da = JSON.parse(daStorage)

            if (da && da.access_token != '') {
                this.authCode = "dummy"
                this.accessToken = da.access_token

                // convert da.expires_at from RFC3339 to Date
                da.expires_at = Date.parse(da.expires_at)

                if (da.expires_at <= Date.now()) {
                    localStorage.removeItem("discord_access")
                    this.authCode = null
                    return
                }
                this.loggedIn = true
            }
        },
        login(code) {
            console.log("LOGGING IN")
            const connectionStore = useConnectionStore()

            connectionStore.addListener("login", (commandResponse) => {
                if (commandResponse.error != '') {
                    if (commandResponse.error == 'invalid_grant') {
                        this.authCode = null
                        // clear the code from params
                        window.history.pushState({}, document.title, window.location.pathname)
                    }
                    return
                }

                this.loggedIn = true
                this.accessToken = commandResponse.result.access_token

                // store the token
                localStorage.setItem("discord_access", JSON.stringify(commandResponse.result))
            })

            connectionStore.send(`login|auth:${code}`)
        },
        getMe() {
            if (this.me) {
                return
            }

            const connectionStore = useConnectionStore()

            connectionStore.addListener("me", (commandResponse) => {
                const errorStore = useErrorStore()

                if (commandResponse.error != '') {
                    if (commandResponse.error != 'user_not_found') {
                        console.error(commandResponse.error)
                    }

                    return
                }

                this.me = new Member(commandResponse.result)

                localStorage.setItem("me", JSON.stringify(commandResponse.result))
            })

            connectionStore.send(`members|me:${this.accessToken}`)
        },
        logout() {
            this.loggedIn = false
            this.accessToken = null
            this.me = null
            localStorage.removeItem("discord_access")
            localStorage.removeItem("me")
        }
    }
})
