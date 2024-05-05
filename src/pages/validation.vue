<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container class="fill-height">
        <v-responsive class="align-centerfill-height mx-auto">

            <v-row class="justify-center">
                <v-col md="6">

                    <v-card class="pa-6 rounded-lg bg-surface-lighten-1">
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-blue.png"
                            v-if="theme.name.value == 'light'" />
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-white.png" v-else />

                        <div class="text-center">

                            <div class="py-2" />

                            Time to validate
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

const theme = useTheme()
const logoColor = ref("blue")

onMounted(() => {
    // if the theme is dark, set the logo to white
    if (theme.name.value == "dark") {
        logoColor.value = "white"
    }
})

</script>
<route lang="yaml">
meta:
  layout: plain
</route>
