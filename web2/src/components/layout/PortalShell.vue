<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { RouterLink, useRoute, useRouter } from 'vue-router'
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

  function isRouteActive (path: string): boolean {
    return route.path === path
  }

  function onNavigate () {
    mobileOpen.value = false
  }

  async function handleLogout () {
    await authStore.logout()
    mobileOpen.value = false
    router.push('/auth/login')
  }

  function buildTopics (): string[] {
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
  <div class="min-h-screen text-on-background">
    <header class="sticky top-0 z-30 border-b border-subtle bg-glass-surface backdrop-blur-12">
      <div class="mx-auto flex h-16 max-w-7xl items-center gap-3 px-4 sm:px-6 lg:px-8">
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-subtle text-on-surface md:hidden"
          type="button"
          @click="mobileOpen = !mobileOpen"
        >
          <span class="sr-only">Toggle navigation</span>
          <span aria-hidden="true" class="text-xl">=</span>
        </button>

        <p class="text-xl font-bold text-primary">Sol Armada</p>
      </div>

      <div
        v-if="realtimeBannerVisible"
        class="border-t border-subtle bg-surface-variant/50 px-4 py-2 text-sm font-medium text-on-surface"
      >
        {{ realtimeBannerText }}
      </div>
    </header>

    <div class="mx-auto flex max-w-7xl gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <aside class="hidden w-64 shrink-0 md:block">
        <nav class="rounded-xl border border-subtle bg-glass-surface p-3">
          <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Member</p>

          <RouterLink
            v-for="item in memberItems"
            :key="item.path"
            class="mt-1 block rounded-md px-3 py-2 text-sm"
            :class="isRouteActive(item.path)
              ? 'bg-primary/20 text-primary'
              : 'text-on-surface hover:bg-surface-variant/40'"
            :to="item.path"
          >
            {{ item.title }}
          </RouterLink>

          <template v-if="adminItems.length > 0">
            <div class="my-3 border-t border-subtle" />
            <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Administration</p>

            <RouterLink
              v-for="item in adminItems"
              :key="item.path"
              class="mt-1 block rounded-md px-3 py-2 text-sm"
              :class="isRouteActive(item.path)
                ? 'bg-secondary/25 text-secondary'
                : 'text-on-surface hover:bg-surface-variant/40'"
              :to="item.path"
            >
              {{ item.title }}
            </RouterLink>
          </template>

          <div class="my-4 border-t border-subtle" />

          <button
            class="w-full rounded-md border border-subtle px-3 py-2 text-sm font-semibold text-on-surface hover:bg-surface-variant/40"
            type="button"
            @click="handleLogout"
          >
            Logout
          </button>
        </nav>
      </aside>

      <main class="min-w-0 flex-1">
        <slot />
      </main>
    </div>

    <div v-if="mobileOpen" class="fixed inset-0 z-40 md:hidden">
      <button class="absolute inset-0 bg-background/75" type="button" @click="mobileOpen = false" />

      <aside class="absolute left-0 top-0 h-full w-72 border-r border-subtle bg-surface p-3">
        <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Member</p>

        <RouterLink
          v-for="item in memberItems"
          :key="`mobile-${item.path}`"
          class="mt-1 block rounded-md px-3 py-2 text-sm"
          :class="isRouteActive(item.path)
            ? 'bg-primary/20 text-primary'
            : 'text-on-surface hover:bg-surface-variant/40'"
          :to="item.path"
          @click="onNavigate"
        >
          {{ item.title }}
        </RouterLink>

        <template v-if="adminItems.length > 0">
          <div class="my-3 border-t border-subtle" />
          <p class="px-2 py-2 text-xs uppercase tracking-wide text-on-surface-variant">Administration</p>

          <RouterLink
            v-for="item in adminItems"
            :key="`mobile-admin-${item.path}`"
            class="mt-1 block rounded-md px-3 py-2 text-sm"
            :class="isRouteActive(item.path)
              ? 'bg-secondary/25 text-secondary'
              : 'text-on-surface hover:bg-surface-variant/40'"
            :to="item.path"
            @click="onNavigate"
          >
            {{ item.title }}
          </RouterLink>
        </template>

        <div class="my-4 border-t border-subtle" />

        <button
          class="w-full rounded-md border border-subtle px-3 py-2 text-sm font-semibold text-on-surface hover:bg-surface-variant/40"
          type="button"
          @click="handleLogout"
        >
          Logout
        </button>
      </aside>
    </div>
  </div>
</template>
