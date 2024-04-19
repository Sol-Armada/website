<template>
    <v-container class="fill-height">
        <v-responsive class="align-centerfill-height mx-auto" max-width="900">

            <v-row class="justify-center py-5">
                <v-col cols="auto">

                    <v-card class="pa-6 rounded-lg">
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-blue.png" />

                        <div class="text-center">

                            <div class="py-2" />

                            <v-btn prepend-icon="fa:fas fa-brands fa-discord" size="large" color="discord-primary"
                                v-if="appStore.authCode == null" :href="discordAuthUrl">Login with Discord</v-btn>

                            <v-progress-circular :size="50" color="primary" indeterminate v-else></v-progress-circular>
                        </div>
                    </v-card>

                </v-col>
            </v-row>
        </v-responsive>
    </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const discordAuthUrl = ref(import.meta.env.VITE_DISCORD_AUTH_URL)

// get the code from the query params
const urlParams = new URLSearchParams(window.location.search)

if (urlParams.has('code')) {
    const code = urlParams.get('code')
    appStore.$patch({ authCode: code })
    appStore.login(code)
}

</script>

<route lang="yaml">
meta:
  layout: login
</route>
