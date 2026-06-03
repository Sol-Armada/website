<script setup lang="ts">
  defineProps<{
    label: string
    value: string | number
    detail?: string
  }>()
</script>

<template>
  <section class="rounded-xl border border-subtle bg-glass-surface p-4">
    <p class="text-xs uppercase tracking-wide text-on-surface-variant">{{ label }}</p>
    <p class="mt-2 text-2xl font-semibold text-on-surface">{{ value }}</p>
    <p v-if="detail" class="mt-2 text-xs text-on-surface-variant">{{ detail }}</p>
  </section>
</template>
