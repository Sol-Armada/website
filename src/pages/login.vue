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
                                v-if="!appStore.loggingIn" :href="discordAuthUrl">Login with Discord</v-btn>

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
import { useTheme } from 'vuetify'

const appStore = useAppStore()

const discordAuthUrl = ref(import.meta.env.VITE_DISCORD_AUTH_URL)
const theme = useTheme()
console.log(theme.name.value)
const logoColor = ref("blue")

// get the code from the query params
const urlParams = new URLSearchParams(window.location.search)

if (urlParams.has('code')) {
    appStore.login(urlParams.get('code'))
}

onMounted(() => {
    // if the theme is dark, set the logo to white
    if (theme.name.value == "dark") {
        logoColor.value = "white"
    }
})

</script>

<route lang="yaml">
meta:
  layout: login
</route>
