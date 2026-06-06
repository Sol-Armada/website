<script setup lang="ts">
  defineProps<{
    title: string
    message: string
    tone?: 'info' | 'error' | 'success'
  }>()
</script>

<template>
  <div
    class="rounded-xl border p-4"
    :class="{
      'border-subtle bg-glass-surface': !tone || tone === 'info',
      'border-error/60 bg-error/8': tone === 'error',
      'border-success/60 bg-success/8': tone === 'success',
    }"
  >
    <p class="text-sm font-semibold text-on-surface">{{ title }}</p>
    <p class="mt-2 text-sm text-on-surface-variant">{{ message }}</p>
  </div>
</template>
