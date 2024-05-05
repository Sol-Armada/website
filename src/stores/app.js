// Utilities
import { defineStore } from 'pinia'
import { useConnectionStore } from './connection'
import { useErrorStore } from './error'
import { Member } from './classes'

export const useAppStore = defineStore('app', () => {
    const errorStore = useErrorStore()
    const connectionStore = useConnectionStore()

    const loggedIn = ref(false)
    const token = ref(null)
    const me = ref(null)
    const onboarded = ref(false)

    const version = ref(null)

    function bindEvents() {
        loggedIn.value = localStorage.getItem('logged_in')
        me.value = JSON.parse(localStorage.getItem('me'))
        token.value = localStorage.getItem('token')
        onboarded.value = localStorage.getItem('onboarded')
        version.value = localStorage.getItem('version')

        connectionStore.addListener('version', 'check').then((commandResponse) => {
            // handle errors
            if (commandResponse.error) {
                errorStore.$patch({ error: commandResponse.error, show: true })
                return
            }
            if (version.value && commandResponse.result != version.value) {
                logout()
            }

            localStorage.setItem('version', commandResponse.result)
        }).catch((error) => {
            if (error == 'invalid_access') {
                logout()
            }
        })

        connectionStore.send('version', 'check')
    }
    async function login(code) {
        return new Promise((resolve, reject) => {
            if (!code) {
                return
            }
            connectionStore.addListener('login', 'auth').then((commandResponse) => {
                // handle errors
                if (commandResponse.error) {
                    if (commandResponse.error == 'invalid_grant') {
                        // remove code from url
                        window.history.replaceState({}, document.title, "/login")
                        reject(false)
                        return
                    }

                    errorStore.$patch({ error: commandResponse.error, show: true })
                    reject(false)
                    return
                }

                // set token
                token.value = commandResponse.result
                loggedIn.value = true

                if (me.value && me.value.onboarded_at) {
                    onboarded.value = true
                }

                save()

                resolve(true)
            }).catch((error) => {
                reject(error)
            })

            connectionStore.send('login', 'auth', code)
        })
    }

    function refresh() { }
    async function getMe() {
        console.log("getting me")
        return new Promise((resolve) => {
            if (me.value) {
                resolve(true)
                return
            }

            connectionStore.addListener('members', 'me').then((commandResponse) => {
                // handle errors
                if (commandResponse.error) {
                    if (commandResponse.error == 'invalid_grant' || commandResponse.error == 'unauthorized') {
                        // remove code from url
                        window.history.replaceState({}, document.title, "/login")
                        resolve(false)
                        return
                    }
                }

                me.value = new Member(commandResponse.result)

                if (me.value && me.value.onboarded_at) {
                    onboarded.value = true
                }

                save()
                resolve(true)
            }).catch((error) => {
                console.log(error)
            })

            connectionStore.send('members', 'me')
        })
    }

    async function checkRSIHandle(handle) {
        return new Promise((resolve) => {
            connectionStore.addListener('rsi', 'check_handle').then((commandResponse) => {
                // handle errors
                if (commandResponse.error) {
                    errorStore.$patch({ error: commandResponse.error, show: true })
                    return
                }

                return resolve(commandResponse.result)
            }).catch((error) => {
                console.log(error)
            })

            connectionStore.send('rsi', 'check_handle', handle)
        })
    }

    async function updateSelf() {
        me.value.onboarded_at = new Date()

        return new Promise((resolve) => {
            connectionStore.addListener('members', 'update-me').then((commandResponse) => {
                // handle errors
                if (commandResponse.error) {
                    errorStore.$patch({ error: commandResponse.error, show: true })
                    return
                }

                resolve(true)
            }).catch((error) => {
                console.log(error)
            })

            connectionStore.send('members', 'update-me', JSON.stringify(me.value))
        })
    }

    function logout() {
        console.log("logging out")
        loggedIn.value = false
        token.value = null
        me.value = null
        onboarded.value = false
        localStorage.removeItem('logged_in')
        localStorage.removeItem('me')
        localStorage.removeItem('token')
        localStorage.removeItem('onboarded')

        // go to login page
        window.location.href = "/login"
    }

    function save() {
        localStorage.setItem("token", token.value)
        localStorage.setItem("me", JSON.stringify(me.value))
        localStorage.setItem("logged_in", loggedIn.value)

        if (me.value && me.value.onboarded_at) {
            localStorage.setItem('onboarded', true)
        }
    }

    return {
        loggedIn,
        token,
        me,
        bindEvents,
        login,
        refresh,
        getMe,
        checkRSIHandle,
        updateSelf,
        logout
    }
})
