<script lang="ts" setup>
  import { onMounted } from 'vue'
  import { useRouter } from 'vue-router'

  const router = useRouter()

  onMounted(() => {
    router.replace('/dashboard')
  })
</script>

<template>
  <section class="mx-auto mt-12 max-w-lg rounded-xl border border-subtle bg-glass-surface p-6 text-center">
    <p class="text-sm text-on-surface-variant">Redirecting to dashboard skeleton...</p>
  </section>
</template>
