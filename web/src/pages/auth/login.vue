<script setup lang="ts">
  import logo from '@/assets/logo.png'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { useAuthStore } from '@/stores/auth'

  const authStore = useAuthStore()

  async function startLogin() {
    await authStore.login()
  }
</script>

<template>
  <div class="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4">
    <section class="w-full max-w-lg rounded-xl border border-subtle bg-glass-surface p-6">
      <div class="mb-6 flex justify-center">
        <img alt="Sol Armada" class="h-100 w-auto" :src="logo">
      </div>

      <button
        class="flex h-11 w-full items-center justify-center gap-2 rounded-md bg-[#5865F2] px-4 text-base font-semibold text-white transition-colors hover:bg-[#4752C4] active:bg-[#3C45A5] disabled:cursor-not-allowed disabled:bg-[#5865F2]/70"
        :disabled="authStore.loading"
        type="button"
        @click="startLogin"
      >
        <svg
          aria-hidden="true"
          class="h-5 w-5"
          fill="currentColor"
          viewBox="0 0 24 24"
        >
          <path d="M20.317 4.369A19.791 19.791 0 0 0 15.409 3c-.21.375-.45.88-.617 1.275a18.27 18.27 0 0 0-5.583 0A13.102 13.102 0 0 0 8.591 3a19.736 19.736 0 0 0-4.91 1.37C.574 9.066-.269 13.646.153 18.162A19.927 19.927 0 0 0 6.13 21c.48-.66.907-1.356 1.279-2.086-.704-.265-1.374-.598-2.005-.99.168-.123.333-.25.494-.381 3.87 1.818 8.067 1.818 11.892 0 .162.132.327.258.494.38a12.72 12.72 0 0 1-2.007.992A13.564 13.564 0 0 0 17.557 21a19.84 19.84 0 0 0 5.98-2.838c.496-5.236-.847-9.774-3.22-13.793ZM8.02 15.331c-1.165 0-2.123-1.078-2.123-2.403 0-1.326.938-2.403 2.122-2.403 1.194 0 2.142 1.088 2.122 2.403 0 1.325-.938 2.403-2.122 2.403Zm7.961 0c-1.165 0-2.122-1.078-2.122-2.403 0-1.326.938-2.403 2.122-2.403 1.194 0 2.142 1.088 2.122 2.403 0 1.325-.928 2.403-2.122 2.403Z" />
        </svg>
        {{ authStore.loading ? 'Redirecting...' : 'Continue with Discord' }}
      </button>

      <div v-if="authStore.error" class="mt-4">
        <StatePanel :message="authStore.error" title="Authentication error" tone="error" />
      </div>
    </section>
  </div>
</template>
