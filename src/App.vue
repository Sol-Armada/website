<template>
    <router-view />
    <Error :text="error" :show="show" :loading="loading" :timeout="timeout" :closable="closable" />
</template>

<script setup>
import { onMounted } from "vue"
import { useAppStore } from "./stores/app"
import { storeToRefs } from "pinia"

import { useConnectionStore } from "@/stores/connection"
import { useErrorStore } from '@/stores/error'

const appStore = useAppStore()
const connectionStore = useConnectionStore()
const errorStore = useErrorStore()

const { loggedIn } = storeToRefs(appStore)
const { error, show, loading, timeout, closable } = storeToRefs(errorStore)

connectionStore.bindEvents()
appStore.bindEvents()

appStore.$subscribe((mutation, state) => {
    if (state.loggedIn && window.location.href.indexOf("/login") != -1) {
        window.location.href = "/"
    }
    if (!state.loggedIn && window.location.href.indexOf("/login") == -1) {
        window.location.href = "/login"
    }
})

onMounted(() => {
    if (!loggedIn.value && window.location.href.indexOf("/login") == -1) {
        window.location.href = "/login"
    } else if (loggedIn.value && window.location.href.indexOf("/login") != -1) {
        window.location.href = "/"
    }
})

function onError(error) {
    errorMsg.value = error
    errorShow.value = true
}
</script>
