<template>
    <v-container fluid class="login-container">
        <v-row align="center" justify="center" class="fill-height">
            <v-col cols="12" sm="8" md="6" lg="4">
                <v-card class="login-card pa-8" elevation="12">
                    <v-card-text class="text-center">
                        <!-- Logo/Title -->
                        <div class="mb-8">
                            <h1 class="text-h3 mb-2 text-primary font-weight-bold">
                                Sol Armada
                            </h1>
                            <p class="text-h6 text-medium-emphasis">
                                Member Dashboard
                            </p>
                        </div>

                        <!-- Welcome Text -->
                        <div class="mb-8">
                            <p class="text-body-1 text-medium-emphasis">
                                Sign in with your Discord account to access the Sol Armada member portal
                            </p>
                        </div>

                        <!-- Login Button -->
                        <v-btn color="primary" size="x-large" block :loading="authStore.loading"
                            :disabled="authStore.loading" @click="handleLogin" class="discord-login-btn">
                            <v-icon left class="mr-2">mdi-discord</v-icon>
                            Sign in with Discord
                        </v-btn>

                        <!-- Error Message -->
                        <v-alert v-if="authStore.error" type="error" variant="tonal" class="mt-4" closable
                            @click:close="authStore.error = null">
                            {{ authStore.error }}
                        </v-alert>

                        <!-- Info -->
                        <div class="mt-8 text-caption text-medium-emphasis">
                            <p>You must be a member of the Sol Armada Discord server</p>
                            <p class="mt-2">Having trouble? Contact an administrator</p>
                        </div>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// Check if already authenticated
onMounted(async () => {
    const isAuth = await authStore.checkAuth()
    if (isAuth) {
        router.push('/dashboard')
    }
})

const handleLogin = async () => {
    await authStore.login()
}
</script>

<style scoped>
.login-container {
    min-height: 100vh;
    background: linear-gradient(135deg, #0a0e27 0%, #141829 100%);
}

.login-card {
    background: rgba(20, 24, 41, 0.95) !important;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(0, 217, 255, 0.1);
}

.discord-login-btn {
    font-weight: 600;
    letter-spacing: 0.5px;
    text-transform: none;
}
</style>
