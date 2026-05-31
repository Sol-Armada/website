<template>
    <div class="overview-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-primary">
                    System Overview
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    High-level view of guild statistics
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

        <!-- Stats Cards -->
        <v-row v-else>
            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Members</div>
                        <div class="text-h3 font-weight-bold text-primary mt-2">
                            {{ stats.totalMembers }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Events</div>
                        <div class="text-h3 font-weight-bold text-accent mt-2">
                            {{ stats.totalEvents }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Tokens Distributed</div>
                        <div class="text-h3 font-weight-bold text-secondary mt-2">
                            {{ stats.totalTokens.toLocaleString() }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Active This Month</div>
                        <div class="text-h3 font-weight-bold text-success mt-2">
                            {{ stats.activeThisMonth }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Unique Attendees (30d)</div>
                        <div class="text-h3 font-weight-bold text-info mt-2">
                            {{ stats.uniqueAttendees }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12" sm="6" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Avg Attendance per Event</div>
                        <div class="text-h3 font-weight-bold text-warning mt-2">
                            {{ stats.averageAttendance }}
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import adminService from '@/services/adminService'

const loading = ref(true)
const error = ref<string | null>(null)
const stats = ref({
    totalMembers: 0,
    totalEvents: 0,
    totalTokens: 0,
    activeThisMonth: 0,
    uniqueAttendees: 0,
    averageAttendance: 0,
})

onMounted(async () => {
    loading.value = true
    error.value = null

    try {
        const response = await adminService.getOverview()
        stats.value = response
    } catch (err: any) {
        error.value = err.message || 'Failed to load overview statistics'
    } finally {
        loading.value = false
    }
})
</script>

<style scoped>
.overview-page {
    max-width: 1400px;
    margin: 0 auto;
}

.stat-card {
    background: rgba(20, 24, 41, 0.8) !important;
    border: 1px solid rgba(0, 217, 255, 0.1);
    height: 100%;
}
</style>
