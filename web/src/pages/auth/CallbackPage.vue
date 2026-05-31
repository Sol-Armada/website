<template>
    <v-container fluid class="callback-container">
        <v-row align="center" justify="center" class="fill-height">
            <v-col cols="12" sm="8" md="6" lg="4">
                <v-card class="callback-card pa-8" elevation="12">
                    <v-card-text class="text-center">
                        <!-- Loading State -->
                        <div v-if="!error">
                            <v-progress-circular indeterminate color="primary" size="64" width="6" class="mb-6" />
                            <h2 class="text-h5 mb-2">Authenticating</h2>
                            <p class="text-body-2 text-medium-emphasis">
                                Please wait while we verify your Discord account...
                            </p>
                        </div>

                        <!-- Error State -->
                        <div v-else>
                            <v-icon size="64" color="error" class="mb-4">
                                mdi-alert-circle-outline
                            </v-icon>
                            <h2 class="text-h5 mb-2 text-error">Authentication Failed</h2>
                            <p class="text-body-1 mb-6 text-medium-emphasis">
                                {{ error }}
                            </p>
                            <v-btn color="primary" size="large" @click="goToLogin">
                                Return to Login
                            </v-btn>
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const error = ref<string | null>(null)

onMounted(async () => {
    // The backend OAuth callback endpoint handles the code exchange
    // It sets the session cookie and redirects here
    // We just need to fetch the user data to populate the store

    try {
        // Check for error in query params (from backend)
        if (route.query.error) {
            error.value = route.query.message as string || 'Authentication failed'
            return
        }

        // Fetch user data (should be authenticated via cookie)
        await authStore.fetchUser()

        // Redirect to dashboard
        router.push('/dashboard')
    } catch (err: any) {
        console.error('Callback error:', err)
        error.value = err.message || 'Failed to complete authentication'
    }
})

const goToLogin = () => {
    router.push('/auth/login')
}
</script>

<style scoped>
.callback-container {
    min-height: 100vh;
    background: linear-gradient(135deg, #0a0e27 0%, #141829 100%);
}

.callback-card {
    background: rgba(20, 24, 41, 0.95) !important;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(0, 217, 255, 0.1);
}
</style>
