<script setup lang="ts">
  import { onMounted, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { useAuthStore } from '@/stores/auth'

  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()
  const callbackError = ref<string | null>(null)

  onMounted(async() => {
    const queryError = route.query.error
    if (typeof queryError === 'string' && queryError.length > 0) {
      const queryMessage = route.query.message
      callbackError.value = typeof queryMessage === 'string' && queryMessage.length > 0
        ? queryMessage
        : 'Authentication failed during callback'
      return
    }

    try {
      await authStore.fetchUser()
      router.replace('/dashboard')
    } catch {
      callbackError.value = authStore.error || 'Failed to restore session after callback'
    }
  })
</script>

<template>
  <section class="mx-auto max-w-lg rounded-xl border border-subtle bg-glass-surface p-6">
    <PageHeader
      subtitle="Completing Discord login and restoring your server session."
      title="Callback"
    />

    <StatePanel
      v-if="!callbackError"
      message="Checking auth cookie and loading your user profile..."
      title="Redirecting"
    />

    <div v-else>
      <StatePanel
        :message="callbackError"
        title="Authentication failed"
        tone="error"
      />

      <button
        class="mt-4 rounded-md border border-subtle px-4 py-2 text-sm font-semibold text-on-surface hover:bg-surface-variant/40"
        type="button"
        @click="router.push('/auth/login')"
      >
        Return to login
      </button>
    </div>
  </section>
</template>
