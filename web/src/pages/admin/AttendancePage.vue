<template>
    <div class="attendance-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-primary">
                    Attendance Records
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    Manage and view event attendance
                </p>
                <v-alert v-if="error" type="error" variant="tonal" class="mt-4" dense>
                    {{ error }}
                </v-alert>
            </v-col>
        </v-row>

        <!-- Loading state -->
        <v-row v-if="loading" class="justify-center py-12">
            <v-progress-circular indeterminate color="primary" size="64" />
        </v-row>

        <!-- Attendance Table -->
        <v-row v-else>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-title class="d-flex align-center">
                        <v-icon class="mr-2">mdi-calendar-check</v-icon>
                        Events
                    </v-card-title>
                    <v-divider />
                    <v-data-table :headers="headers" :items="records" :loading="loading" class="attendance-table">
                        <template #item.dateCreated="{ item }">
                            {{ new Date(item.dateCreated).toLocaleDateString() }}
                        </template>
                        <template #item.recorded="{ item }">
                            <v-chip :color="item.recorded ? 'success' : 'warning'" size="small">
                                {{ item.recorded ? 'Recorded' : 'Pending' }}
                            </v-chip>
                        </template>
                        <template #item.successful="{ item }">
                            <v-icon :color="item.successful ? 'success' : 'warning'">
                                {{ item.successful ? 'mdi-check-circle' : 'mdi-alert-circle' }}
                            </v-icon>
                        </template>
                    </v-data-table>
                    <v-card-actions class="justify-center pa-4">
                        <v-pagination v-model="page" :length="totalPages" />
                    </v-card-actions>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import adminService, { type AttendanceRecord } from '@/services/adminService'

const loading = ref(true)
const error = ref<string | null>(null)
const records = ref<AttendanceRecord[]>([])
const page = ref(1)
const limit = ref(50)

const headers = [
    { title: 'Name', key: 'name' },
    { title: 'Submitted By', key: 'submittedBy' },
    { title: 'Participants', key: 'participantCount' },
    { title: 'Status', key: 'recorded' },
    { title: 'Successful', key: 'successful' },
    { title: 'Date', key: 'dateCreated' },
]

const totalPages = computed(() => Math.ceil(records.value.length / limit.value))

const loadAttendance = async () => {
    loading.value = true
    error.value = null

    try {
        const response = await adminService.getAttendance(limit.value, page.value)
        records.value = response.records || []
    } catch (err: any) {
        error.value = err.message || 'Failed to load attendance records'
    } finally {
        loading.value = false
    }
}

onMounted(loadAttendance)
</script>

<style scoped>
.attendance-page {
    max-width: 1400px;
    margin: 0 auto;
}

.attendance-table {
    background: rgba(20, 24, 41, 0.8) !important;
}
</style>
