<template>
    <v-app>
        <Navigation v-if="!isMobile()" :member="appStore.me" :logout="logout" />

        <v-app-bar title="Application bar" v-else></v-app-bar>

        <v-main class="d-flex justify-center scrollable" style="min-height: 300px;">
            <router-view />
        </v-main>
    </v-app>
</template>

<script setup>
import { useAppStore } from '@/stores/app'

const emit = defineEmits(['onLogout'])

const appStore = useAppStore()

onMounted(() => {
    if (appStore.accessToken != null) {
        appStore.getMe()
    }
})

function isMobile() {
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

function logout() {
    appStore.logout()
}
</script>

<style lang="scss"></style>
