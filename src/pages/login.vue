<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container class="fill-height">
        <v-responsive class="align-centerfill-height mx-auto" max-width="900">

            <v-row class="justify-center py-5">
                <v-col cols="auto">

                    <v-card class="pa-6 rounded-lg bg-surface-lighten-1">
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-blue.png"
                            v-if="theme.name.value == 'light'" />
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-white.png" v-else />

                        <div class="text-center">

                            <div class="py-2" />

                            <v-btn prepend-icon="fa:fas fa-brands fa-discord" size="large" color="discord-primary"
                                :disabled="appStore.loggingIn || !connectionStore.isConnected"
                                v-if="!appStore.loggingIn && code === null" :href="discordAuthUrl">Login with
                                Discord</v-btn>

                            <v-progress-circular :size="50" color="primary" indeterminate v-else></v-progress-circular>
                        </div>
                    </v-card>

                </v-col>
            </v-row>
        </v-responsive>
    </v-container>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { useConnectionStore } from '@/stores/connection'
import { useTheme } from 'vuetify'

const appStore = useAppStore()
const connectionStore = useConnectionStore()

const discordAuthUrl = ref(import.meta.env.VITE_DISCORD_AUTH_URL)
const theme = useTheme()
const logoColor = ref("blue")
const code = ref(new URLSearchParams(window.location.search).get('code'))

onMounted(() => {
    // if the theme is dark, set the logo to white
    if (theme.name.value == "dark") {
        logoColor.value = "white"
    }

    if (code.value) {
        appStore.login(code.value).then((res) => {
            if (res) {
                console.log("LOGGED IN")
                appStore.getMe().then(() => {
                    console.log("GOT ME")
                    if (appStore.me && appStore.me.onboarded) {
                        window.location.href = "/"
                    } else {
                        window.location.href = "/onboard"
                    }
                })
            }
        })
    }
})

</script>

<route lang="yaml">
meta:
  layout: plain
</route>
