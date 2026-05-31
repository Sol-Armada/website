<template>
    <v-app>
        <!-- App Bar -->
        <v-app-bar color="surface" elevation="2" app>
            <v-app-bar-nav-icon @click="drawer = !drawer" />

            <v-toolbar-title class="text-primary font-weight-bold">
                Sol Armada
            </v-toolbar-title>

            <v-spacer />

            <!-- User Menu -->
            <v-menu offset-y>
                <template v-slot:activator="{ props }">
                    <v-btn v-bind="props" icon size="large">
                        <v-avatar v-if="authStore.user?.avatar" size="32">
                            <v-img :src="authStore.user.avatar" />
                        </v-avatar>
                        <v-icon v-else>mdi-account-circle</v-icon>
                    </v-btn>
                </template>

                <v-list>
                    <v-list-item>
                        <v-list-item-title class="font-weight-bold">
                            {{ authStore.user?.username }}
                        </v-list-item-title>
                        <v-list-item-subtitle class="text-caption">
                            {{ authStore.user?.email }}
                        </v-list-item-subtitle>
                    </v-list-item>

                    <v-divider />

                    <v-list-item prepend-icon="mdi-account" @click="goToProfile">
                        <v-list-item-title>Profile</v-list-item-title>
                    </v-list-item>

                    <v-list-item prepend-icon="mdi-logout" @click="handleLogout">
                        <v-list-item-title>Logout</v-list-item-title>
                    </v-list-item>
                </v-list>
            </v-menu>
        </v-app-bar>

        <!-- Navigation Drawer -->
        <v-navigation-drawer v-model="drawer" app color="surface">
            <v-list>
                <!-- Member Section -->
                <v-list-subheader>MEMBER</v-list-subheader>

                <v-list-item v-for="item in memberItems" :key="item.path" :to="item.path" :prepend-icon="item.icon"
                    :title="item.title" color="primary" />

                <!-- Admin Section (only for admin/moderator) -->
                <template v-if="authStore.hasAnyRole(['admin', 'moderator'])">
                    <v-divider class="my-2" />
                    <v-list-subheader>ADMINISTRATION</v-list-subheader>

                    <v-list-item v-for="item in adminItems" :key="item.path" :to="item.path" :prepend-icon="item.icon"
                        :title="item.title" color="secondary"
                        :disabled="item.requiresAdmin && !authStore.hasRole('admin')" />
                </template>
            </v-list>

            <template v-slot:append>
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
import { ref, computed } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const drawer = ref(true)

const memberItems = computed(() => [
    {
        title: 'Dashboard',
        icon: 'mdi-view-dashboard',
        path: '/dashboard',
    },
    {
        title: 'Profile',
        icon: 'mdi-account',
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
        requiresAdmin: false, // Moderators can access
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

const goToProfile = () => {
    router.push('/dashboard/profile')
}

const handleLogout = async () => {
    await authStore.logout()
    router.push('/auth/login')
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
