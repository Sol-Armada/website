<template>
    <v-container fluid>
        <v-row>
            <v-col cols="12">
                <v-card class="bg-card-on-surface">
                    <v-card-title>Members</v-card-title>
                    <v-divider></v-divider>

                    <v-card class="bg-card-on-surface">
                        <template v-slot:text>
                            <v-text-field v-model="search" label="Search" prepend-inner-icon="mdi:mdi-magnify"
                                variant="outlined" hide-details single-line></v-text-field>
                        </template>
                        <v-container fluid :style="{ height: '100%' }">
                            <v-data-table class="bg-card-on-surface" :items="members" :disable-items-per-page=true
                                :headers="headers" density="compact" :search="search" :loading="loading"
                                :itemsPerPageOptions="[12]" :loading-text="loadingText" color="white"
                                v-model:page="page" v-touch="{
                                    left: () => swipe('Left'),
                                    right: () => swipe('Right'),
                                    up: () => swipe('Up'),
                                    down: () => swipe('Down')
                                }"></v-data-table>
                        </v-container>
                    </v-card>

                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
<script setup>
import Member from '@/components/member.vue'

import { computed, onMounted, ref } from 'vue'
import { useTheme } from 'vuetify'
import { useMembersStore } from '@/stores/members'

const memberStore = useMembersStore()

const loading = ref(true)
const loadingText = ref('Loading members... 0')

const headers = [{ title: 'Name', key: 'name' }, { title: 'Rank', key: 'rank.name' }, { title: 'Events Attended', key: 'eventsAttended' }]
const members = ref([])
const membersPage = ref(0)
const page = ref(1)
const search = ref('')

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}

const pageCount = computed(() => {
    return Math.ceil(members.value.length / 10)
})

function swipe(direction) {
    if (direction == 'Left') {
        if (page.value < pageCount.value) {
            page.value += 1
        }
    } else if (direction == 'Right') {
        if (page.value > 1) {
            page.value -= 1
        }
    }
}

onMounted(async () => {
    let m = []
    while (true) {
        let moreMembers = await memberStore.getMembers(membersPage.value)
        if (moreMembers.length == 0) {
            console.log("no more members")
            loading.value = false
            break
        }
        m = m.concat(moreMembers)
        loadingText.value = `Loading members... (${m.length})`
        membersPage.value += 1
    }
    members.value = m
})

async function load() {
    page.value += 1
    const moreMembers = await memberStore.getMembers(page.value)
    members.value = members.value.concat(moreMembers)
}

</script>
<style>
.v-data-table-footer__items-per-page {
    display: none;
}
</style>
<route lang="yaml">
meta:
    layout: default
    requiresOfficer: true
</route>
