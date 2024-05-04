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

    function bindEvents() {
        loggedIn.value = localStorage.getItem('logged_in')
        me.value = JSON.parse(localStorage.getItem('me'))
        token.value = localStorage.getItem('token')

        // if (me.value == null) {
        //     getMe()
        // }
    }
    function login(code) {
        connectionStore.addListener('login', 'auth', (commandResponse) => {
            // handle errors
            if (commandResponse.error) {
                if (commandResponse.error == 'invalid_grant') {
                    console.log("invalid_grant")
                    // remove code from url
                    window.history.replaceState({}, document.title, "/login")
                    return
                }

                errorStore.$patch({ error: commandResponse.error, show: true })
                return
            }

            // set token
            token.value = commandResponse.result
            loggedIn.value = true

            save()
        })

        connectionStore.send('login', 'auth', code)
    }

    function refresh() { }
    function getMe() {
        if (me.value) {
            return
        }

        connectionStore.addListener('members', 'me', (commandResponse) => {
            // handle errors
            if (commandResponse.error) {
                if (commandResponse.error == 'invalid_grant') {
                    console.log("invalid_grant")
                    // remove code from url
                    window.history.replaceState({}, document.title, "/login")
                    return
                }
            }

            console.log(commandResponse)

            me.value = new Member(commandResponse.result)

            save()
        })

        console.log("getting me")

        connectionStore.send('members', 'me')
    }

    function logout() {
        loggedIn.value = false
        token.value = null
        me.value = null
        localStorage.removeItem('logged_in')
        localStorage.removeItem('me')
        localStorage.removeItem('token')
    }

    function save() {
        localStorage.setItem("token", token.value)
        localStorage.setItem("me", JSON.stringify(me.value))
        localStorage.setItem("logged_in", loggedIn.value)
    }

    return {
        loggedIn,
        token,
        me,
        bindEvents,
        login,
        refresh,
        getMe,
        logout
    }
})
