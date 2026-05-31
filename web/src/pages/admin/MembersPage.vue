<template>
    <div class="members-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-secondary">
                    Member Management
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View and manage organization members
                </p>
            </v-col>
        </v-row>

        <!-- Search and Filters -->
        <v-row class="mb-4">
            <v-col cols="12" md="6">
                <v-text-field v-model="search" prepend-inner-icon="mdi-magnify" label="Search members"
                    variant="outlined" density="compact" clearable />
            </v-col>
            <v-col cols="12" md="6">
                <v-select v-model="roleFilter" :items="['All', 'Admin', 'Moderator', 'Member']" label="Filter by role"
                    variant="outlined" density="compact" />
            </v-col>
        </v-row>

        <!-- Members Table Placeholder -->
        <v-row>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-text class="text-center py-12">
                        <v-icon size="64" color="secondary" class="mb-4">mdi-account-group</v-icon>
                        <p class="text-h6">Member management coming in Phase 6</p>
                        <p class="text-body-2 text-medium-emphasis mt-2">
                            Integration with sol-bot member data
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
const roleFilter = ref('All')
</script>

<style scoped>
.members-page {
    max-width: 1400px;
    margin: 0 auto;
}
</style>
