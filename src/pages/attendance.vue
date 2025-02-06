<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container fluid>
        <v-row class="justify-center">
            <v-col cols="8">
                <v-card class="bg-card-on-surface">
                    <v-card-title>Attendance Records</v-card-title>
                    <v-divider></v-divider>

                    <v-card class="bg-card-on-surface">
                        <template v-slot:text>
                            <v-text-field v-model="search" label="Search" prepend-inner-icon="mdi:mdi-magnify"
                                variant="outlined" hide-details single-line></v-text-field>
                        </template>
                        <v-container fluid :style="{ height: '100%' }">
                            <v-data-table id="attendance-records" class="bg-card-on-surface" :items="attendanceRecords"
                                :disable-items-per-page=true :headers="headers" density="compact" :search="search"
                                :loading="loading" :itemsPerPageOptions="[12]" :loading-text="loadingText" color="white"
                                v-model:page="page" v-touch="{
                                    left: () => swipe('Left'),
                                    right: () => swipe('Right')
                                }" v-model:sort-by.sync="sortBy">
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
                                                    <v-card-text>{{ item.numberOfMembers }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ new Date(item.createdDate).toLocaleString()
                                                        }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.submittedBy.name }}</v-card-text>
                                                </v-card>
                                            </v-col>
                                        </td>
                                        <td>
                                            <v-col cols="12">
                                                <v-card>
                                                    <v-card-text>{{ item.recorded ? 'Yes' : 'No' }}</v-card-text>
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
import { useAttendanceStore } from '@/stores/attendance'
import { storeToRefs } from 'pinia'

const attendanceStore = useAttendanceStore()
const { attendance } = storeToRefs(attendanceStore)

const loading = ref(true)
const loadingText = ref('Loading attendance records... 0')

const headers = [
    { title: 'Name', key: 'name' },
    { title: 'Member Count', key: 'memberCount' },
    { title: 'Created Date', key: 'dateCreated' },
    { title: 'Submitted by', key: 'submittedBy' },
    { title: 'Recorded', key: 'recorded' },
]
const attendanceRecords = ref(attendance.value)
const attendanceRecordsPage = ref(1)
const page = ref(1)
const search = ref('')
const sortBy = ref([{ key: 'date_created', order: 'desc' }]);

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}

const pageCount = computed(() => {
    return Math.ceil(attendanceRecords.value.length / 10)
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
    attendanceRecordsPage.value = 1
    // eslint-disable-next-line no-constant-condition
    while (true) {
        loadingText.value = `Loading attendance records... (${attendanceRecords.value.length})`
        const moreRecords = await attendanceStore.getAttendanceRecords(attendanceRecordsPage.value)
        if (moreRecords.length == 0) {
            console.debug("no more attendance records")
            loading.value = false
            break
        }
        attendanceRecordsPage.value += 1
    }
})

</script>
<style lang="scss">
#attendance-records .v-data-table-footer__items-per-page {
    display: none !important;
}

#attendance-records tr>td {
    border: none;
}

#attendance-records tr>td:first-of-type {
    padding-right: 0;

    .v-col {
        padding: 5px 0 5px 0;
    }

    .v-card {
        border-radius: 8px 0 0 8px;
    }
}

#attendance-records tr>td:not(:first-of-type):not(:last-of-type) {
    padding-right: 0;
    padding-left: 0;

    .v-col {
        padding: 0;
    }

    .v-card {
        border-radius: 0;
    }
}

#attendance-records tr>td:last-of-type {
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
