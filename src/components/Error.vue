<template>
    <v-snackbar v-model="show" color="error" location="top" :timeout="timeout">
        {{ error }}

        <template v-slot:actions v-if="closable">
            <v-btn variant="text" @click="show = false">
                Close
            </v-btn>
        </template>
        <v-progress-linear indeterminate v-if="loading"></v-progress-linear>
    </v-snackbar>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useErrorStore } from '@/stores/error'

const errorStore = useErrorStore()
const { error, show, loading, timeout, closable } = storeToRefs(errorStore)
</script>
