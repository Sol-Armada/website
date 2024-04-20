// Utilities
import { defineStore } from 'pinia'
import { useConnectionStore } from './connection'
import { useErrorStore } from './error'
import { Member, CommandResponse } from './classes'

export const useAppStore = defineStore('app', {
    state: () => ({
        loggedIn: false,
        loggingIn: false,
        /** @type {string} */
        token: null,
        /** @type {string} */
        me: null
    }),
    actions: {
        bindEvents() {
            const meStorage = localStorage.getItem("me")

            if (meStorage) {
                this.me = new Member(JSON.parse(meStorage))
            }

            this.token = localStorage.getItem("token")

            if (this.token == null || this.token == "null") {
                this.logout()
                return
            }

            const loggedInStorage = localStorage.getItem("logged_in")

            if (loggedInStorage == "true") {
                this.loggedIn = true
            }
        },
        login(code) {
            const connectionStore = useConnectionStore()

            connectionStore.addListener("login", (commandResponse) => {
                if (commandResponse.action == 'refresh') {
                    if (commandResponse.error == 'invalid_access') {
                        this.logout()
                        return
                    }

                    localStorage.setItem("token", commandResponse.result)
                }

                if (commandResponse.action == 'auth') {
                    if (commandResponse.error != '') {
                        if (commandResponse.error == 'invalid_grant' || commandResponse.error == 'invalid_access') {
                            this.logout()
                            return
                        }

                        console.log("LOGIN ERROR: " + commandResponse.error)
                        useErrorStore().$patch({ error: "Ran into a server error. Please try again later.", show: true, timeout: 5000, closeable: false, reason: "login" })
                        return
                    }

                    this.loggedIn = true
                    this.loggingIn = false
                    this.token = commandResponse.result.token

                    // store the token
                    localStorage.setItem("token", commandResponse.result)
                    localStorage.setItem("logged_in", "true")
                }
            })

            connectionStore.send('login', 'auth', code)
            this.loggingIn = true
        },
        refresh() {
            useConnectionStore().send(`login|refresh:`)
        },
        getMe() {
            if (this.me) {
                return
            }
            const connectionStore = useConnectionStore()

            connectionStore.addListener("members", (commandResponse) => {
                if (commandResponse.action == 'me') {
                    if (commandResponse.error != '') {
                        if (commandResponse.error != 'user_not_found') {
                            console.error(commandResponse.error)
                        }
                        this.logout()
                        return
                    }

                    this.me = new Member(commandResponse.result)

                    localStorage.setItem("me", JSON.stringify(commandResponse.result))
                }
            })
            connectionStore.send('members', 'me', '')
        },
        logout() {
            this.loggedIn = false
            this.token = null
            this.me = null
            localStorage.removeItem("token")
            localStorage.removeItem("me")
            localStorage.removeItem("logged_in")
            // remove code from query
            window.history.replaceState({}, document.title, window.location.pathname)
        }
    }
})
