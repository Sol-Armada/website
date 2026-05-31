<template>
    <div class="token-ledger-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-primary">
                    Token Ledger
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View all token transactions and distributions
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

        <!-- Token Table -->
        <v-row v-else>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-title class="d-flex align-center">
                        <v-icon class="mr-2">mdi-currency-usd</v-icon>
                        Transactions
                    </v-card-title>
                    <v-divider />
                    <v-data-table :headers="headers" :items="records" :loading="loading" class="token-table">
                        <template #item.amount="{ item }">
                            <span :class="item.amount > 0 ? 'text-success' : 'text-warning'">
                                {{ item.amount > 0 ? '+' : '' }}{{ item.amount }}
                            </span>
                        </template>
                        <template #item.createdAt="{ item }">
                            {{ new Date(item.createdAt).toLocaleDateString() }}
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
import adminService, { type TokenTransaction } from '@/services/adminService'

const loading = ref(true)
const error = ref<string | null>(null)
const records = ref<TokenTransaction[]>([])
const page = ref(1)
const limit = ref(50)

const headers = [
    { title: 'Member ID', key: 'memberId' },
    { title: 'Amount', key: 'amount' },
    { title: 'Reason', key: 'reason' },
    { title: 'Comment', key: 'comment' },
    { title: 'Date', key: 'createdAt' },
]

const totalPages = computed(() => Math.ceil(records.value.length / limit.value))

const loadTokenLedger = async () => {
    loading.value = true
    error.value = null

    try {
        const response = await adminService.getTokenLedger(limit.value, page.value)
        records.value = response.records || []
    } catch (err: any) {
        error.value = err.message || 'Failed to load token ledger'
    } finally {
        loading.value = false
    }
}

onMounted(loadTokenLedger)
</script>

<style scoped>
.token-ledger-page {
    max-width: 1400px;
    margin: 0 auto;
}

.token-table {
    background: rgba(20, 24, 41, 0.8) !important;
}
</style>
