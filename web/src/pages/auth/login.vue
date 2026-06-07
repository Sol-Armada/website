<script setup lang="ts">
import PageHeader from '@/components/ui/PageHeader.vue'
import StatePanel from '@/components/ui/StatePanel.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

async function startLogin() {
  await authStore.login()
}
</script>

<template>
  <section class="mx-auto max-w-lg rounded-xl border border-subtle bg-glass-surface p-6">
    <PageHeader subtitle="Use Discord to authenticate. The backend handles OAuth and secure session cookies."
      title="Sign In" />

    <button class="w-full rounded-md bg-primary px-4 py-2 text-sm font-semibold text-on-primary"
      :disabled="authStore.loading" type="button" @click="startLogin">
      {{ authStore.loading ? 'Redirecting...' : 'Sign in with Discord' }}
    </button>

    <div v-if="authStore.error" class="mt-4">
      <StatePanel :message="authStore.error" title="Authentication error" tone="error" />
    </div>
  </section>
</template>
