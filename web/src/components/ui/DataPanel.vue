<script setup lang="ts">
  defineProps<{
    title: string
    description: string
  }>()
</script>

<template>
  <section class="tactical-panel p-5 mb-7">
    <h2 class="text-lg font-semibold text-on-surface">{{ title }}</h2>
    <p class="mt-2 text-sm text-on-surface-variant">{{ description }}</p>

    <div class="mt-4">
      <slot />
    </div>
  </section>
</template>
