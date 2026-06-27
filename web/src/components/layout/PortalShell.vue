<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { RouterLink, useRoute, useRouter } from 'vue-router'
  import logo from '@/assets/mqatt0az-logo.png'
  import { WS_TOPIC_ADMIN_ATTENDANCE, WS_TOPIC_ADMIN_MEMBERS, WS_TOPIC_ADMIN_TOKEN_LEDGER, WS_TOPIC_SYSTEM_HEALTH, wsClient } from '@/services/wsClient'
  import { useAppStore } from '@/stores/app'
  import { type Role, useAuthStore } from '@/stores/auth'

  interface NavItem {
    title: string
    path: string
    roles?: Role[]
  }

  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()
  const appStore = useAppStore()
  const mobileOpen = ref(false)
  const portalVersion = __APP_VERSION__

  const realtimeBannerVisible = computed(() => {
    return appStore.realtimeState === 'connecting'
      || appStore.realtimeState === 'reconnecting'
      || appStore.realtimeState === 'disconnected'
  })

  const realtimeBannerText = computed(() => {
    if (appStore.realtimeState === 'connecting') {
      return 'Connecting to server for realtime updates...'
    }
    if (appStore.realtimeState === 'reconnecting') {
      return 'Server connection lost. Reconnecting...'
    }
    return 'Server cannot be reached right now. Waiting for connection...'
  })

  const memberItems: NavItem[] = [
    { title: 'Dashboard', path: '/dashboard' },
    { title: 'Profile', path: '/dashboard/profile' },
  ]

  const adminItems = computed<NavItem[]>(() => {
    const items: NavItem[] = [
      { title: 'Overview', path: '/admin/overview', roles: ['admin'] },
      { title: 'Attendance', path: '/admin/attendance', roles: ['moderator', 'admin'] },
      { title: 'Token Ledger', path: '/admin/token-ledger', roles: ['admin'] },
      { title: 'Members', path: '/admin/members', roles: ['admin'] },
    ]

    return items.filter(item => {
      if (!item.roles) {
        return true
      }
      return authStore.hasAnyRole(item.roles)
    })
  })

  function isRouteActive(path: string): boolean {
    return route.path === path
  }

  function onNavigate() {
    mobileOpen.value = false
  }

  async function handleLogout() {
    await authStore.logout()
    mobileOpen.value = false
    router.push('/auth/login')
  }

  function buildTopics(): string[] {
    const topics = [WS_TOPIC_SYSTEM_HEALTH]
    if (authStore.hasRole('admin')) {
      topics.push(
        WS_TOPIC_ADMIN_MEMBERS,
        WS_TOPIC_ADMIN_ATTENDANCE,
        WS_TOPIC_ADMIN_TOKEN_LEDGER,
      )
    }
    return topics
  }

  onMounted(() => {
    wsClient.connect(buildTopics())
  })

  watch(
    () => authStore.user?.roles.join('|') || '',
    () => {
      wsClient.connect(buildTopics())
    },
  )
</script>

<template>
  <div class="flex min-h-dvh flex-col text-on-background">
    <!-- Tactical Top Navigation Bar -->
    <header class="sticky top-0 z-30 border-b border-divider bg-surface/95 backdrop-blur-12">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <!-- Logo + Brand -->
        <RouterLink class="flex items-center gap-3" to="/dashboard">
          <img alt="Sol Armada" class="logo-pulse h-9 w-9" :src="logo">
          <span class="text-xl font-bold text-primary">SOL ARMADA</span>
        </RouterLink>

        <!-- Desktop Navigation Links -->
        <div class="hidden items-center gap-1 md:flex">
          <!-- Member Links -->
          <RouterLink
            v-for="item in memberItems"
            :key="item.path"
            class="nav-link"
            :class="{ 'nav-link-active': isRouteActive(item.path) }"
            :to="item.path"
          >
            {{ item.title }}
          </RouterLink>

          <!-- Admin Links (if authorized) -->
          <template v-if="adminItems.length > 0">
            <div class="mx-2 h-6 w-px bg-divider" />

            <RouterLink
              v-for="item in adminItems"
              :key="item.path"
              class="nav-link nav-link-admin"
              :class="{ 'nav-link-admin-active': isRouteActive(item.path) }"
              :to="item.path"
            >
              {{ item.title }}
            </RouterLink>

          </template>

          <!-- Logout Button -->
          <button
            class="ml-4 rounded-md border border-divider px-3 py-2 text-sm font-semibold text-on-surface hover:bg-surface-variant/40 hover:border-primary/50 transition-all"
            type="button"
            @click="handleLogout"
          >
            Logout
          </button>
        </div>

        <!-- Mobile Menu Toggle -->
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-divider text-on-surface md:hidden"
          type="button"
          @click="mobileOpen = !mobileOpen"
        >
          <span class="sr-only">Toggle navigation</span>
          <span aria-hidden="true" class="text-xl">☰</span>
        </button>
      </nav>

      <!-- Realtime Connection Banner -->
      <div
        v-if="realtimeBannerVisible"
        class="border-t border-divider bg-surface-variant/50 px-4 py-2 text-sm font-medium text-on-surface"
      >
        {{ realtimeBannerText }}
      </div>
    </header>

    <!-- Main Content -->
    <main class="mx-auto w-full max-w-7xl flex-1 px-4 py-6 sm:px-6 lg:px-8">
      <slot />
    </main>

    <!-- Footer with Version -->
    <footer class="mt-auto border-t border-divider bg-surface/50 py-4">
      <div class="mx-auto max-w-7xl px-4 text-center text-xs text-on-surface-variant mono-numeric sm:px-6 lg:px-8">
        SOL ARMADA Portal {{ portalVersion }}
      </div>
    </footer>

    <!-- Mobile Menu Drawer -->
    <div v-if="mobileOpen" class="fixed inset-0 z-40 md:hidden">
      <button class="absolute inset-0 bg-background/75" type="button" @click="mobileOpen = false" />

      <aside class="absolute left-0 top-0 flex h-full w-72 flex-col border-r border-divider bg-surface p-4">
        <div class="mb-6 flex items-center gap-3">
          <img alt="Sol Armada" class="h-8 w-8" :src="logo">
          <span class="text-lg font-bold text-primary">SOL ARMADA</span>
        </div>

        <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Member</p>

        <RouterLink
          v-for="item in memberItems"
          :key="`mobile-${item.path}`"
          class="mt-1 block rounded-md px-3 py-2 text-sm transition-colors"
          :class="isRouteActive(item.path)
            ? 'bg-primary/20 text-primary font-semibold'
            : 'text-on-surface hover:bg-surface-variant/40'"
          :to="item.path"
          @click="onNavigate"
        >
          {{ item.title }}
        </RouterLink>

        <template v-if="adminItems.length > 0">
          <div class="my-3 border-t border-divider" />
          <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Administration</p>

          <RouterLink
            v-for="item in adminItems"
            :key="`mobile-admin-${item.path}`"
            class="mt-1 block rounded-md px-3 py-2 text-sm transition-colors"
            :class="isRouteActive(item.path)
              ? 'bg-secondary/25 text-secondary font-semibold'
              : 'text-on-surface hover:bg-surface-variant/40'"
            :to="item.path"
            @click="onNavigate"
          >
            {{ item.title }}
          </RouterLink>
        </template>

        <div class="my-4 border-t border-divider" />

        <button
          class="w-full rounded-md border border-divider px-3 py-2 text-sm font-semibold text-on-surface hover:bg-surface-variant/40 transition-colors"
          type="button"
          @click="handleLogout"
        >
          Logout
        </button>

        <footer class="mt-auto px-2 py-3 text-xs text-on-surface-variant mono-numeric">
          {{ portalVersion }}
        </footer>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.nav-link {
  position: relative;
  display: inline-block;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--v-on-surface);
  text-decoration: none;
  border-radius: 0.375rem;
  transition: all 150ms cubic-bezier(0.2, 0, 0, 1);
}

.nav-link:hover {
  background-color: rgba(230, 168, 45, 0.1);
  color: var(--v-primary);
}

.nav-link-active {
  background-color: rgba(230, 168, 45, 0.15);
  color: var(--v-primary);
  font-weight: 600;
}

.nav-link-admin:hover {
  background-color: rgba(0, 65, 255, 0.1);
  color: var(--v-secondary);
}

.nav-link-admin-active {
  background-color: rgba(0, 65, 255, 0.15);
  color: var(--v-secondary);
  font-weight: 600;
}

.nav-link-admin-active::after {
  background: linear-gradient(90deg, #0041FF 0%, #E6A82D 50%, transparent 100%);
}
</style>
