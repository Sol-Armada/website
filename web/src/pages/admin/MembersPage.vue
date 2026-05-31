<template>
    <div class="members-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12" md="8">
                <h1 class="text-h4 font-weight-bold text-primary">
                    Member Directory
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View and manage guild members
                </p>
            </v-col>
            <v-col cols="12" md="4">
                <v-text-field v-model="search" label="Search members" prepend-inner-icon="mdi-magnify"
                    variant="outlined" density="compact" @update:model-value="resetPage" />
            </v-col>
        </v-row>

        <v-alert v-if="error" type="error" variant="tonal" class="mb-4" dense>
            {{ error }}
        </v-alert>

        <!-- Loading state -->
        <v-row v-if="loading" class="justify-center py-12">
            <v-progress-circular indeterminate color="primary" size="64" />
        </v-row>

        <!-- Members Table -->
        <v-row v-else>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-title class="d-flex align-center">
                        <v-icon class="mr-2">mdi-account-multiple</v-icon>
                        Members
                    </v-card-title>
                    <v-divider />
                    <v-data-table :headers="headers" :items="members" :loading="loading" class="members-table">
                        <template #item.rank="{ item }">
                            <v-chip :color="getRankColor(item.rank)" size="small">
                                {{ item.rank }}
                            </v-chip>
                        </template>
                        <template #item.tokenBalance="{ item }">
                            <span class="font-weight-bold">{{ item.tokenBalance.toLocaleString() }}</span>
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
import adminService, { type MemberSummary } from '@/services/adminService'

const loading = ref(true)
const error = ref<string | null>(null)
const members = ref<MemberSummary[]>([])
const page = ref(1)
const search = ref('')
const limit = ref(50)

const headers = [
    { title: 'Username', key: 'username' },
    { title: 'Rank', key: 'rank' },
    { title: 'Attendance', key: 'attendance' },
    { title: 'Token Balance', key: 'tokenBalance' },
    { title: 'RSI Handle', key: 'rsiHandle' },
]

const totalPages = computed(() => Math.ceil(members.value.length / limit.value))

const getRankColor = (rank: string) => {
    const colors: Record<string, string> = {
        Admiral: 'error',
        Commander: 'warning',
        Lieutenant: 'info',
        Specialist: 'success',
        Technician: 'primary',
        Member: 'secondary',
        Recruit: 'grey',
    }
    return colors[rank] || 'grey'
}

const resetPage = () => {
    page.value = 1
    loadMembers()
}

const loadMembers = async () => {
    loading.value = true
    error.value = null

    try {
        const response = await adminService.getMembers(limit.value, page.value, search.value)
        members.value = response.members || []
    } catch (err: any) {
        error.value = err.message || 'Failed to load members'
    } finally {
        loading.value = false
    }
}

onMounted(loadMembers)
</script>

<style scoped>
.members-page {
    max-width: 1400px;
    margin: 0 auto;
}

.members-table {
    background: rgba(20, 24, 41, 0.8) !important;
}
</style>
