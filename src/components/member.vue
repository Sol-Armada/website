<template>
    <v-col cols="2">
        <v-card class="text-center" min-width="100">
            <v-card-title>{{ member.name }}</v-card-title>
            <v-card-subtitle>{{ member.rank.name }}</v-card-subtitle>
            <v-card-text>Events: {{ member.eventsAttended }}</v-card-text>
        </v-card>
    </v-col>
</template>
<script setup>
import { Member } from '../stores/classes'

const props = defineProps({
    member: Member
})
</script>
