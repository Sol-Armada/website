<template>
    <v-container fluid>
        <v-row>
            <v-col cols="12">
                <v-card-title>members</v-card-title>
                <v-divider></v-divider>
                <v-infinite-scroll :items="members" @load="load">
                    <template v-for="(member, _) in members" :key="member.id">
                        <Member :member="member" />
                    </template>
                </v-infinite-scroll>
                <!-- <v-virtual-scroll :items="members">

                    <template v-slot:default="{ item }">
                        <v-card :class="'bg-surface-' + ld + '-1'">
                            <v-row class="pa-2">
                                <v-col cols="2">
                                    <Member :member="item" />
                                </v-col>
                            </v-row>
                        </v-card>
                    </template>

</v-virtual-scroll> -->
            </v-col>
        </v-row>
    </v-container>
</template>
<script setup>
import Member from '@/components/member.vue'

import { onMounted, ref } from 'vue'
import { useTheme } from 'vuetify'
import { storeToRefs } from 'pinia'
import { useMembersStore } from '@/stores/members'

const memberStore = useMembersStore()

const members = ref([])
const page = ref(1)

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}

onMounted(async () => {
    const m = await memberStore.getMembers(page.value)
    members.value = m
})

async function load({ done }) {
    page.value += 1
    console.log("NEXT PAGE", page.value)
    const moreMembers = await memberStore.getMembers(page.value)
    console.log(moreMembers)
    members.value = members.value.concat(moreMembers)
    done('ok')
}

</script>
<route lang="yaml">
meta:
    layout: default
</route>
