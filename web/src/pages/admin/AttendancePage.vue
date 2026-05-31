<template>
    <div class="attendance-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-secondary">
                    Event Attendance
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    Track and manage member attendance at events
                </p>
            </v-col>
        </v-row>

        <!-- Filters and Actions -->
        <v-row class="mb-4">
            <v-col cols="12" md="6">
                <v-text-field v-model="search" prepend-inner-icon="mdi-magnify" label="Search events or members"
                    variant="outlined" density="compact" clearable />
            </v-col>
            <v-col cols="12" md="6" class="text-right">
                <v-btn color="secondary" prepend-icon="mdi-plus">
                    Record Attendance
                </v-btn>
            </v-col>
        </v-row>

        <!-- Attendance Table Placeholder -->
        <v-row>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-text class="text-center py-12">
                        <v-icon size="64" color="secondary" class="mb-4">mdi-calendar-check</v-icon>
                        <p class="text-h6">Attendance tracking coming in Phase 6</p>
                        <p class="text-body-2 text-medium-emphasis mt-2">
                            Integration with sol-bot attendance data
                        </p>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const search = ref('')
</script>

<style scoped>
.attendance-page {
    max-width: 1400px;
    margin: 0 auto;
}
</style>
