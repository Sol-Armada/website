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
                            <v-text-field v-model="membersListSearch" label="Search"
                                prepend-inner-icon="mdi:mdi-magnify" variant="outlined" hide-details
                                single-line></v-text-field>
                        </template>
                        <v-container fluid :style="{ height: '100%' }">
                            <v-data-table id="members" class="bg-card-on-surface" :items="sortedMembers"
                                :disable-items-per-page=true :headers="headers" density="compact"
                                :search="membersListSearch" :loading="loading" :itemsPerPageOptions="[12]"
                                :loading-text="loadingText" color="white" v-model:page="membersListPage" v-touch="{
                                    left: () => swipe('Left'),
                                    right: () => swipe('Right')
                                }">
                                <template v-slot:item="{ item }">
                                    <tr :id="item.id" @click="memberDetail = item">
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
            <v-col>
                <v-card class="bg-card-on-surface" max-height="100%" max-width="100%">
                    <v-card-title>Recorded Attendance for the Member (In Development)</v-card-title>
                    <v-divider></v-divider>
                    <v-card class="bg-card-on-surface" v-if="memberDetail">
                        <template v-slot:text>
                            <v-text-field v-model="attendanceListSearch" label="Search"
                                prepend-inner-icon="mdi:mdi-magnify" variant="outlined" hide-details
                                single-line></v-text-field>
                        </template>
                        <v-container fluid :style="{ height: '100%' }">
                            <v-data-table id="attendance" class="bg-card-on-surface" :items="attendanceRecords"
                                :disable-items-per-page=true :headers="attendanceHeaders" density="compact"
                                :search="attendanceListSearch" :loading="loading" :itemsPerPageOptions="[12]"
                                :loading-text="loadingText" color="white" v-model:page="attendanceListPage" v-touch="{
                                    left: () => swipe('Left'),
                                    right: () => swipe('Right')
                                }">
                                <template v-slot:item="{ item }">
                                    <tr :id="item.id">
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.name }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.dateCreated }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                    </tr>
                                </template>
                            </v-data-table>
                        </v-container>
                    </v-card>
                    <h3 class="text-center" v-else>
                        Select a member to see their details
                    </h3>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
    <!-- <v-overlay v-model="overlay" class="align-center justify-center">
        <v-card class="bg-card-on-surface" max-height="100%" max-width="100%">
            <v-card-title>Member Details</v-card-title>
            <v-divider></v-divider>
            <v-card class="bg-card-on-surface">
                <template v-slot:text>
                    <v-text-field v-model="search" label="Search" prepend-inner-icon="mdi:mdi-magnify"
                        variant="outlined" hide-details single-line></v-text-field>
                </template>
                <v-container fluid :style="{ height: '100%' }">
                    <v-data-table id="attendance" class="bg-card-on-surface" :items="attendanceRecords"
                        :disable-items-per-page=true :headers="attendanceHeaders" density="compact" :search="search"
                        :loading="loading" :itemsPerPageOptions="[12]" :loading-text="loadingText" color="white"
                        v-model:page="page" v-touch="{
                            left: () => swipe('Left'),
                            right: () => swipe('Right')
                        }">
                        <template v-slot:item="{ item }">
                            <tr :id="item.id">
                                <td>
                                    <v-col cols="12">
                                        <v-card>
                                            <v-card-text>{{ item.name }}</v-card-text>
                                        </v-card>
                                    </v-col>
                                </td>
                            </tr>
                        </template>
                    </v-data-table>
                </v-container>
            </v-card>
        </v-card>
    </v-overlay> -->
</template>
<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useTheme } from 'vuetify'
import { useMembersStore } from '@/stores/members'
import { useAttendanceStore } from '@/stores/attendance';

const memberStore = useMembersStore()
const attendanceStore = useAttendanceStore()

const loading = ref(true)
const loadingText = ref('Loading... 0')

const memberDetail = ref(null)
const overlay = ref(false)

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

const attendanceHeaders = [{ title: 'Name', key: 'name' }, { title: 'Date', key: 'dateCreated' }]
const attendanceRecords = ref([])

const membersListPage = ref(1)
const membersListSearch = ref('')

const attendanceListPage = ref(1)
const attendanceListSearch = ref('')

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
        if (membersListPage.value < pageCount.value) {
            membersListPage.value += 1
        }
    } else if (direction == 'Right') {
        if (membersListPage.value > 1) {
            membersListPage.value -= 1
        }
    }
}

watch(memberDetail, () => {
    if (memberDetail.value) {
        overlay.value = true
        loading.value = true
        attendanceStore.getMemberAttendanceRecords(memberDetail.value.id).then((result) => {
            console.log(result)
            attendanceRecords.value = result

            loading.value = false
        }).catch((err) => {
            console.error(err)
        })
    }
})

onMounted(async () => {
    let membersPage = 1
    // eslint-disable-next-line no-constant-condition
    while (true) {
        loadingText.value = `Loading members... (${members.value.length})`
        const moreMembers = await memberStore.getMembers(membersPage)
        if (moreMembers.length == 0) {
            console.debug("no more members")
            loading.value = false
            break
        }
        members.value.push(...moreMembers)
        membersPage += 1
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

#attendance .v-data-table-footer__items-per-page {
    display: none !important;
}

#attendance tr>td {
    border: none;
}

#attendance tr>td:first-of-type {
    padding-right: 0;

    .v-col {
        padding: 5px 0 5px 0;
    }

    .v-card {
        border-radius: 8px 0 0 8px;
    }
}

#attendance tr>td:not(:first-of-type):not(:last-of-type) {
    padding-right: 0;
    padding-left: 0;

    .v-col {
        padding: 0;
    }

    .v-card {
        border-radius: 0;
    }
}

#attendance tr>td:last-of-type {
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
