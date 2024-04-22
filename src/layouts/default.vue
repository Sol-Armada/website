<template>
    <v-app>
        <Navigation :member="appStore.me" :logout="logout" />

        <v-main class="d-flex justify-center" style="min-height: 300px;">
            <router-view />
        </v-main>
    </v-app>
</template>

<script setup>
import { useAppStore } from '@/stores/app'

const emit = defineEmits(['onLogout'])

const appStore = useAppStore()

onMounted(() => {
    if (!appStore.me) {
        appStore.getMe()
    }
})

function logout() {
    appStore.logout()
}
</script>

<style lang="scss"></style>
