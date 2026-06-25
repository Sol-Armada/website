<script setup lang="ts">
  defineProps<{
    title: string
    subtitle: string
  }>()
</script>

<template>
  <header class="mb-6">
    <h1 class="text-3xl font-bold text-primary">{{ title }}</h1>
    <p v-if="subtitle" class="mt-2 text-sm text-on-surface-variant">{{ subtitle }}</p>
  </header>
</template>
