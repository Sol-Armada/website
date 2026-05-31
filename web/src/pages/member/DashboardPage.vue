<template>
    <div class="dashboard-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-primary">
                    Welcome back, {{ authStore.user?.username }}
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View your attendance, earned tokens, and rank
                </p>
            </v-col>
        </v-row>

        <!-- Stats Cards -->
        <v-row>
            <!-- Attendance Card -->
            <v-col cols="12" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="d-flex align-center mb-2">
                            <v-icon size="32" color="primary" class="mr-3">
                                mdi-calendar-check
                            </v-icon>
                            <div>
                                <div class="text-caption text-medium-emphasis">
                                    Total Attendance
                                </div>
                                <div class="text-h4 font-weight-bold">
                                    {{ loading ? '...' : stats.attendance }}
                                </div>
                            </div>
                        </div>
                        <v-progress-linear :model-value="attendanceProgress" color="primary" height="6" rounded
                            class="mt-4" />
                        <div class="text-caption text-medium-emphasis mt-2">
                            {{ stats.attendance }} events attended
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <!-- Tokens Card -->
            <v-col cols="12" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="d-flex align-center mb-2">
                            <v-icon size="32" color="accent" class="mr-3">
                                mdi-currency-usd
                            </v-icon>
                            <div>
                                <div class="text-caption text-medium-emphasis">
                                    Earned Tokens
                                </div>
                                <div class="text-h4 font-weight-bold">
                                    {{ loading ? '...' : stats.tokens.toLocaleString() }}
                                </div>
                            </div>
                        </div>
                        <v-progress-linear :model-value="tokenProgress" color="accent" height="6" rounded
                            class="mt-4" />
                        <div class="text-caption text-medium-emphasis mt-2">
                            Lifetime earnings
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>

            <!-- Rank Card -->
            <v-col cols="12" md="4">
                <v-card elevation="2" class="stat-card">
                    <v-card-text>
                        <div class="d-flex align-center mb-2">
                            <v-icon size="32" color="secondary" class="mr-3">
                                mdi-shield-star
                            </v-icon>
                            <div>
                                <div class="text-caption text-medium-emphasis">
                                    Current Rank
                                </div>
                                <div class="text-h4 font-weight-bold">
                                    {{ loading ? '...' : stats.rank }}
                                </div>
                            </div>
                        </div>
                        <v-chip :color="getRankColor(stats.rank)" variant="flat" class="mt-4">
                            {{ getRankTitle(stats.rank) }}
                        </v-chip>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>

        <!-- Recent Activity -->
        <v-row class="mt-4">
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-title class="d-flex align-center">
                        <v-icon class="mr-2">mdi-history</v-icon>
                        Recent Activity
                    </v-card-title>
                    <v-divider />
                    <v-card-text>
                        <v-list v-if="!loading && recentActivity.length > 0" lines="two">
                            <v-list-item v-for="(activity, index) in recentActivity" :key="index">
                                <template v-slot:prepend>
                                    <v-avatar :color="getActivityColor(activity.type)">
                                        <v-icon>{{ getActivityIcon(activity.type) }}</v-icon>
                                    </v-avatar>
                                </template>
                                <v-list-item-title>{{ activity.title }}</v-list-item-title>
                                <v-list-item-subtitle>{{ activity.date }}</v-list-item-subtitle>
                            </v-list-item>
                        </v-list>
                        <div v-else-if="loading" class="text-center py-8">
                            <v-progress-circular indeterminate color="primary" />
                        </div>
                        <div v-else class="text-center py-8 text-medium-emphasis">
                            <v-icon size="48" class="mb-2">mdi-information-outline</v-icon>
                            <p>No recent activity</p>
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(true)

// Mock stats - will be replaced with API calls in Phase 5
const stats = ref({
    attendance: 0,
    tokens: 0,
    rank: 'N/A',
})

const recentActivity = ref<Array<{ type: string, title: string, date: string }>>([])

const attendanceProgress = computed(() => {
    // Mock progress calculation
    return Math.min((stats.value.attendance / 50) * 100, 100)
})

const tokenProgress = computed(() => {
    // Mock progress calculation
    return Math.min((stats.value.tokens / 10000) * 100, 100)
})

const getRankColor = (rank: string) => {
    const ranks: Record<string, string> = {
        'Recruit': 'grey',
        'Member': 'blue',
        'Veteran': 'green',
        'Elite': 'purple',
        'Legend': 'orange',
    }
    return ranks[rank] || 'grey'
}

const getRankTitle = (rank: string) => {
    return rank || 'Unranked'
}

const getActivityColor = (type: string) => {
    const colors: Record<string, string> = {
        'attendance': 'primary',
        'token': 'accent',
        'rank': 'secondary',
    }
    return colors[type] || 'grey'
}

const getActivityIcon = (type: string) => {
    const icons: Record<string, string> = {
        'attendance': 'mdi-calendar-check',
        'token': 'mdi-currency-usd',
        'rank': 'mdi-shield-star',
    }
    return icons[type] || 'mdi-information'
}

onMounted(async () => {
    // TODO: Fetch member stats from API in Phase 5
    // For now, show mock data
    setTimeout(() => {
        stats.value = {
            attendance: 0,
            tokens: 0,
            rank: 'Recruit',
        }
        loading.value = false
    }, 500)
})
</script>

<style scoped>
.dashboard-page {
    max-width: 1400px;
    margin: 0 auto;
}

.stat-card {
    background: rgba(20, 24, 41, 0.8) !important;
    border: 1px solid rgba(0, 217, 255, 0.1);
    height: 100%;
}
</style>
