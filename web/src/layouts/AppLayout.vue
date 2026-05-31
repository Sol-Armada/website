<template>
    <v-app>
        <!-- App Bar -->
        <v-app-bar color="surface" elevation="2" app>
            <v-app-bar-nav-icon v-if="isMobile" @click="drawer = !drawer" />

            <v-toolbar-title class="text-primary font-weight-bold">
                Sol Armada
            </v-toolbar-title>

            <v-spacer />
        </v-app-bar>

        <!-- Navigation Drawer -->
        <v-navigation-drawer v-model="drawer" app color="surface" :permanent="!isMobile" :temporary="isMobile">
            <v-list>
                <!-- Member Section -->
                <v-list-subheader>MEMBER</v-list-subheader>

                <v-list-item v-for="item in memberItems" :key="item.path" :to="item.path" :prepend-icon="item.icon"
                    :title="item.title" color="primary" :active="isRouteActive(item.path)" />

                <!-- Admin Section (only for admin/moderator) -->
                <template v-if="authStore.hasAnyRole(['admin', 'moderator'])">
                    <v-divider class="my-2" />
                    <v-list-subheader>ADMINISTRATION</v-list-subheader>

                    <v-list-item v-for="item in adminItems" :key="item.path" :to="item.path" :prepend-icon="item.icon"
                        :title="item.title" color="secondary" :active="isRouteActive(item.path)"
                        :disabled="item.requiresAdmin && !authStore.hasRole('admin')" />
                </template>
            </v-list>

            <template v-slot:append>
                <v-divider />
                <div class="pa-2">
                    <v-btn block variant="text" prepend-icon="mdi-logout" :loading="authStore.loading"
                        @click="handleLogout">
                        Logout
                    </v-btn>
                </div>
                <v-divider />
                <div class="pa-4 text-center text-caption text-medium-emphasis">
                    <p>v1.0.0</p>
                    <p class="mt-1">Sol Armada © 2026</p>
                </div>
            </template>
        </v-navigation-drawer>

        <!-- Main Content -->
        <v-main>
            <v-container fluid>
                <RouterView />
            </v-container>
        </v-main>
    </v-app>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDisplay } from 'vuetify'

const router = useRouter()
const authStore = useAuthStore()
const drawer = ref(true)
const { mobile } = useDisplay()
const isMobile = computed(() => mobile.value)

watch(isMobile, (value) => {
    if (!value) {
        drawer.value = true
    }
}, { immediate: true })

const memberItems = computed(() => [
    {
        title: 'Dashboard',
        icon: 'mdi-view-dashboard',
        path: '/dashboard',
    },
    {
        title: 'Profile',
        icon: 'mdi-account-circle',
        path: '/dashboard/profile',
    },
])

const adminItems = computed(() => [
    {
        title: 'Overview',
        icon: 'mdi-chart-box',
        path: '/admin/overview',
        requiresAdmin: true,
    },
    {
        title: 'Attendance',
        icon: 'mdi-calendar-check',
        path: '/admin/attendance',
        requiresAdmin: false,
    },
    {
        title: 'Token Ledger',
        icon: 'mdi-currency-usd',
        path: '/admin/token-ledger',
        requiresAdmin: true,
    },
    {
        title: 'Members',
        icon: 'mdi-account-group',
        path: '/admin/members',
        requiresAdmin: true,
    },
])

const handleLogout = async () => {
    await authStore.logout()
    router.push('/auth/login')
}

const isRouteActive = (path: string): boolean => {
    return router.currentRoute.value.path === path
}
</script>

<style scoped>
.v-navigation-drawer {
    border-right: 1px solid rgba(0, 217, 255, 0.1);
}

.v-app-bar {
    border-bottom: 1px solid rgba(0, 217, 255, 0.1);
}
</style>
