<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-app>
        <Navigation v-if="me.isOfficer" />

        <v-main class="d-flex justify-center" style="min-height: 300px;">
            <router-view />
        </v-main>
    </v-app>
</template>

<script setup>
import { useAppStore } from '@/stores/app'
import { storeToRefs } from 'pinia'

const appStore = useAppStore()

const { me } = storeToRefs(appStore)

</script>

<style lang="scss"></style>
