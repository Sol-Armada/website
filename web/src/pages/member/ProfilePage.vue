<template>
    <div class="profile-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-primary">
                    My Profile
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View and manage your account information
                </p>
            </v-col>
        </v-row>

        <!-- Profile Card -->
        <v-row>
            <v-col cols="12" md="8">
                <v-card elevation="2" class="profile-card">
                    <v-card-text>
                        <!-- Avatar and Basic Info -->
                        <div class="d-flex align-center mb-6">
                            <v-avatar size="100" class="mr-6">
                                <v-img v-if="authStore.user?.avatar" :src="authStore.user.avatar" />
                                <v-icon v-else size="64">mdi-account-circle</v-icon>
                            </v-avatar>
                            <div>
                                <h2 class="text-h5 font-weight-bold">
                                    {{ authStore.user?.username }}
                                </h2>
                                <p class="text-body-2 text-medium-emphasis">
                                    {{ authStore.user?.email }}
                                </p>
                                <div class="mt-2">
                                    <v-chip v-for="role in authStore.user?.roles" :key="role"
                                        :color="getRoleColor(role)" size="small" class="mr-2">
                                        {{ role }}
                                    </v-chip>
                                </div>
                            </div>
                        </div>

                        <v-divider class="my-4" />

                        <!-- Account Information -->
                        <h3 class="text-h6 mb-4">Account Information</h3>

                        <v-list>
                            <v-list-item>
                                <template v-slot:prepend>
                                    <v-icon>mdi-identifier</v-icon>
                                </template>
                                <v-list-item-title>User ID</v-list-item-title>
                                <v-list-item-subtitle>{{ authStore.user?.id }}</v-list-item-subtitle>
                            </v-list-item>

                            <v-list-item>
                                <template v-slot:prepend>
                                    <v-icon>mdi-discord</v-icon>
                                </template>
                                <v-list-item-title>Discord ID</v-list-item-title>
                                <v-list-item-subtitle>{{ authStore.user?.discordID }}</v-list-item-subtitle>
                            </v-list-item>

                            <v-list-item>
                                <template v-slot:prepend>
                                    <v-icon>mdi-shield-check</v-icon>
                                </template>
                                <v-list-item-title>Roles</v-list-item-title>
                                <v-list-item-subtitle>{{ authStore.user?.roles.join(', ') }}</v-list-item-subtitle>
                            </v-list-item>
                        </v-list>
                    </v-card-text>
                </v-card>
            </v-col>

            <!-- Quick Stats -->
            <v-col cols="12" md="4">
                <v-card elevation="2" class="mb-4">
                    <v-card-title>
                        <v-icon class="mr-2">mdi-chart-line</v-icon>
                        Quick Stats
                    </v-card-title>
                    <v-divider />
                    <v-card-text>
                        <v-alert v-if="error" type="error" variant="tonal" dense>
                            {{ error }}
                        </v-alert>
                        <div v-else-if="loading" class="text-center py-4">
                            <v-progress-circular indeterminate color="primary" />
                        </div>
                        <v-list v-else density="compact" lines="one">
                            <v-list-item>
                                <v-list-item-title>Rank</v-list-item-title>
                                <template v-slot:append>
                                    <v-chip color="secondary" size="small">{{ profileStats.rank }}</v-chip>
                                </template>
                            </v-list-item>
                            <v-list-item>
                                <v-list-item-title>Attendance</v-list-item-title>
                                <template v-slot:append>
                                    <strong>{{ profileStats.attendanceCount }}</strong>
                                </template>
                            </v-list-item>
                            <v-list-item>
                                <v-list-item-title>Token Balance</v-list-item-title>
                                <template v-slot:append>
                                    <strong>{{ profileStats.tokensBalance.toLocaleString() }}</strong>
                                </template>
                            </v-list-item>
                            <v-list-item v-if="profileStats.memberSince">
                                <v-list-item-title>Member Since</v-list-item-title>
                                <template v-slot:append>
                                    <span>{{ profileStats.memberSince }}</span>
                                </template>
                            </v-list-item>
                            <v-list-item v-if="profileStats.rsiHandle">
                                <v-list-item-title>RSI Handle</v-list-item-title>
                                <template v-slot:append>
                                    <span>{{ profileStats.rsiHandle }}</span>
                                </template>
                            </v-list-item>
                        </v-list>
                    </v-card-text>
                </v-card>

                <v-card elevation="2">
                    <v-card-title>
                        <v-icon class="mr-2">mdi-information</v-icon>
                        Account Status
                    </v-card-title>
                    <v-divider />
                    <v-card-text>
                        <v-chip color="success" variant="flat" block>
                            <v-icon left>mdi-check-circle</v-icon>
                            Active
                        </v-chip>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import memberService from '@/services/memberService'

const authStore = useAuthStore()
const loading = ref(true)
const error = ref<string | null>(null)
const profileStats = ref({
    rank: 'Recruit',
    attendanceCount: 0,
    tokensBalance: 0,
    memberSince: '',
    rsiHandle: '',
})

const getRoleColor = (role: string) => {
    const colors: Record<string, string> = {
        'admin': 'error',
        'moderator': 'warning',
        'member': 'primary',
    }
    return colors[role] || 'grey'
}

onMounted(async () => {
    loading.value = true
    error.value = null

    try {
        const profile = await memberService.getProfile()
        profileStats.value = {
            rank: profile.rank,
            attendanceCount: profile.attendanceCount,
            tokensBalance: profile.tokensBalance,
            memberSince: profile.memberSince ? new Date(profile.memberSince).toLocaleDateString() : '',
            rsiHandle: profile.rsiHandle || '',
        }
    } catch (err: any) {
        error.value = err.message || 'Failed to load profile statistics'
    } finally {
        loading.value = false
    }
})
</script>

<style scoped>
.profile-page {
    max-width: 1400px;
    margin: 0 auto;
}

.profile-card {
    background: rgba(20, 24, 41, 0.8) !important;
    border: 1px solid rgba(0, 217, 255, 0.1);
}
</style>
