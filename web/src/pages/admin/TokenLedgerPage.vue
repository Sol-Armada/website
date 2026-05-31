<template>
    <div class="token-ledger-page">
        <!-- Header -->
        <v-row class="mb-6">
            <v-col cols="12">
                <h1 class="text-h4 font-weight-bold text-secondary">
                    Token Ledger
                </h1>
                <p class="text-body-1 text-medium-emphasis mt-2">
                    View and manage token distributions
                </p>
            </v-col>
        </v-row>

        <!-- Summary Cards -->
        <v-row class="mb-4">
            <v-col cols="12" md="4">
                <v-card elevation="2" class="token-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Total Distributed</div>
                        <div class="text-h4 font-weight-bold mt-2">0</div>
                    </v-card-text>
                </v-card>
            </v-col>
            <v-col cols="12" md="4">
                <v-card elevation="2" class="token-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">This Month</div>
                        <div class="text-h4 font-weight-bold mt-2">0</div>
                    </v-card-text>
                </v-card>
            </v-col>
            <v-col cols="12" md="4">
                <v-card elevation="2" class="token-card">
                    <v-card-text>
                        <div class="text-caption text-medium-emphasis">Average per Member</div>
                        <div class="text-h4 font-weight-bold mt-2">0</div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>

        <!-- Ledger Table Placeholder -->
        <v-row>
            <v-col cols="12">
                <v-card elevation="2">
                    <v-card-text class="text-center py-12">
                        <v-icon size="64" color="accent" class="mb-4">mdi-currency-usd</v-icon>
                        <p class="text-h6">Token ledger coming in Phase 6</p>
                        <p class="text-body-2 text-medium-emphasis mt-2">
                            Integration with sol-bot token system
                        </p>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.token-ledger-page {
    max-width: 1400px;
    margin: 0 auto;
}

.token-card {
    background: rgba(20, 24, 41, 0.8) !important;
    border: 1px solid rgba(255, 159, 28, 0.1);
}
</style>
