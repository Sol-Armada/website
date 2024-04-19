<template>
        <v-snackbar v-model="props.show" color="error" location="top" :timeout="props.timeout">
            {{ text }}

            <template v-slot:actions v-if="props.closable">
                <v-btn variant="text" @click="props.show = false">
                    Close
                </v-btn>
            </template>
            <v-progress-linear indeterminate v-if="props.loading"></v-progress-linear>
        </v-snackbar>
</template>

<script setup>
const props = defineProps({
    text: String,
    show: Boolean,
    timeout: Number,
    loading: Boolean,
    closable: Boolean
})
</script>
