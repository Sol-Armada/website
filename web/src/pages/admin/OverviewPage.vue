<template>
    <div class="admin-overview-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-secondary">
                    Admin Overview
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    System-wide statistics and insights
                </p>
            </v-col>
        </v-row>

        <!-- Stats Cards -->
        <v-row>
            <v-col cols="12" sm="6" md="3">
                <v-card elevation="2" class="admin-stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Members</div>
                        <div class="text-h4 font-weight-bold mt-2">{{ stats.totalMembers }}</div>
                    </v-card-text>
                </v-card>
            </v-col>
            <v-col cols="12" sm="6" md="3">
                <v-card elevation="2" class="admin-stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Events</div>
                        <div class="text-h4 font-weight-bold mt-2">{{ stats.totalEvents }}</div>
                    </v-card-text>
                </v-card>
            </v-col>
            <v-col cols="12" sm="6" md="3">
                <v-card elevation="2" class="admin-stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Tokens Distributed</div>
                        <div class="text-h4 font-weight-bold mt-2">{{ stats.totalTokens.toLocaleString() }}</div>
                    </v-card-text>
                </v-card>
            </v-col>
            <v-col cols="12" sm="6" md="3">
                <v-card elevation="2" class="admin-stat-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Active This Month</div>
                        <div class="text-h4 font-weight-bold mt-2">{{ stats.activeThisMonth }}</div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>

        <!-- Placeholder for charts -->
        <v-row class="mt-4">
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-title>
                        <v-icon class="mr-2">mdi-chart-line</v-icon>
                        Activity Charts
                    </v-card-title>
                    <v-divider />
                    <v-card-text class="text-center py-12">
                        <v-icon size="64" color="secondary" class="mb-4">mdi-chart-areaspline</v-icon>
                        <p class="text-h6">Charts and analytics coming in Phase 6</p>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const stats = ref({
    totalMembers: 0,
    totalEvents: 0,
    totalTokens: 0,
    activeThisMonth: 0,
})
</script>

<style scoped>
.admin-overview-page {
    max-width: 1400px;
    margin: 0 auto;
}

.admin-stat-card {
    background: rgba(20, 24, 41, 0.8) !important;
    border: 1px solid rgba(255, 0, 110, 0.1);
}
</style>
