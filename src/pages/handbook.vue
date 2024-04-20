<template>
    <v-container fluid>
        <v-row>
            <v-col cols="12">
                <v-card :class="'bg-surface-' + ld + '-1'">
                    <v-card-title>Handbook</v-card-title>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
<script setup>
import { ref } from 'vue'
import { useTheme } from 'vuetify'

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}
</script>
<route lang="yaml">
meta:
    layout: default
</route>
