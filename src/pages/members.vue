<template>
    <v-container fluid v-scroll.self="onScroll">
        <v-row>
            <v-col cols="12">
                <v-card :class="'bg-surface-' + ld + '-1'">
                    <v-card-title>members</v-card-title>
                    <v-divider></v-divider>
                    <v-row class="pa-2">
                        <Member v-for="member in members" :member="member" :key="member.id" />
                    </v-row>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
<script setup>
import Member from '@/components/member.vue'

import { ref } from 'vue'
import { useTheme } from 'vuetify'
import { storeToRefs } from 'pinia'
import { useMembersStore } from '@/stores/members'

const memberStore = useMembersStore()

const { members } = storeToRefs(memberStore)

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}

memberStore.bindEvents()

function onScroll() {
    console.log("TEST")
}

</script>
<route lang="yaml">
meta:
    layout: default
</route>
