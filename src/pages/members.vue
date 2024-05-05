<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container fluid>
        <v-row class="justify-center">
            <v-col cols="6">
                <v-card class="bg-card-on-surface">
                    <v-card-title>Members</v-card-title>
                    <v-divider></v-divider>

                    <v-card class="bg-card-on-surface">
                        <template v-slot:text>
                            <v-text-field v-model="search" label="Search" prepend-inner-icon="mdi:mdi-magnify"
                                variant="outlined" hide-details single-line></v-text-field>
                        </template>
                        <v-container fluid :style="{ height: '100%' }">
                            <v-data-table id="members" class="bg-card-on-surface" :items="sortedMembers"
                                :disable-items-per-page=true :headers="headers" density="compact" :search="search"
                                :loading="loading" :itemsPerPageOptions="[12]" :loading-text="loadingText" color="white"
                                v-model:page="page" v-touch="{
                                    left: () => swipe('Left'),
                                    right: () => swipe('Right')
                                }">
                                <template v-slot:item="{ item }">
                                    <tr :id="item.id">
                                        <td width="33.33%">
                                            <v-col cols="12">
                                                <v-card :border="item.rank.color + ' s-xl'">
                                                    <v-card-text>{{ item.name }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td width="25%">
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ !item.isMember ? item.affiliation :
                                                        item.rank.name }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.eventsAttended }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.validated ? 'Yes' : 'No' }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                    </tr>
                                </template>
                            </v-data-table>
                        </v-container>
                    </v-card>

                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useTheme } from 'vuetify'
import { useMembersStore } from '@/stores/members'

const memberStore = useMembersStore()

const loading = ref(true)
const loadingText = ref('Loading members... 0')

const headers = [{ title: 'Name', key: 'name' }, { title: 'Rank', key: 'rank.name' }, { title: 'Events Attended', key: 'eventsAttended' }, { title: "Validated", key: "validated" }]
const members = ref([])
// sort members with rank 0 being on bottom
const sortedMembers = computed(() => {
    let m = members.value
    return m.sort((a, b) => {
        if (a.rank.id < b.rank.id) {
            if (a.rank.id == 0) {
                return 1
            }

            return -1
        }

        if (a.rank.id > b.rank.id) {
            if (b.rank.id == 0) {
                return -1
            }

            return 1
        }

        if (a.rank.id == b.rank.id) {
            // affiliate
            if (a.isAffiliate) {
                if (!b.isAffiliate) {
                    return 1
                }

                return -1
            }

            if (b.isAffiliate) {
                if (!a.isAffiliate) {
                    return -1
                }

                return 1
            }

            if (a.isAffiliate && b.isAffiliate) {
                return 0
            }

            // ally
            if (a.isAlly) {
                if (!b.isAlly) {
                    return 1
                }

                return -1
            }

            if (b.isAlly) {
                if (!a.isAlly) {
                    return -1
                }

                return 1
            }

            if (a.isAlly && b.isAlly) {
                return 0
            }
        }

        if (a.name < b.name) {
            return -1
        }

        if (a.name > b.name) {
            return 1
        }

        return 0
    })
})
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
    membersPage.value = 1
    // eslint-disable-next-line no-constant-condition
    while (true) {
        loadingText.value = `Loading members... (${members.value.length})`
        const moreMembers = await memberStore.getMembers(membersPage.value)
        if (moreMembers.length == 0) {
            console.log("no more members")
            loading.value = false
            break
        }
        members.value.push(...moreMembers)
        membersPage.value += 1
    }
})

</script>
<style lang="scss">
#members .v-data-table-footer__items-per-page {
    display: none !important;
}

#members tr>td {
    border: none;
}

#members tr>td:first-of-type {
    padding-right: 0;

    .v-col {
        padding: 5px 0 5px 0;
    }

    .v-card {
        border-radius: 8px 0 0 8px;
    }
}

#members tr>td:not(:first-of-type):not(:last-of-type) {
    padding-right: 0;
    padding-left: 0;

    .v-col {
        padding: 0;
    }

    .v-card {
        border-radius: 0;
    }
}

#members tr>td:last-of-type {
    padding-left: 0;

    .v-col {
        padding: 0;
    }

    .v-card {
        border-radius: 0 8px 8px 0;
    }
}
</style>
<route lang="yaml">
meta:
    layout: default
    requiresOfficer: true
</route>
