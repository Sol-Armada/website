<template>
    <router-view />
    <Error />
</template>

<script setup>
import { useAppStore } from "@/stores/app"
import { useConnectionStore } from "@/stores/connection"

const appStore = useAppStore()
const connectionStore = useConnectionStore()

connectionStore.bindEvents()
appStore.bindEvents()

appStore.$subscribe((mutation, state) => {
    if (!state.loggedIn || !state.token) {
        // redirect to login if not logged in
        if (window.location.pathname != "/login") {
            window.location.href = "/login"
        }
    }

    if (state.loggedIn && state.token) {
        // redirect to home if already logged in
        if (window.location.pathname == "/login") {
            window.location.href = "/"
        }
    }
})
</script>
