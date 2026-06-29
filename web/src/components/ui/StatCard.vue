<script setup lang="ts">
  defineProps<{
    label: string
    value: string | number
    detail?: string
    colorBorder?: boolean
  }>()
</script>

<template>
  <section :class="['sweep-animation p-4', { 'tactical-panel': colorBorder, 'panel': !colorBorder }]">
    <p class="text-xs uppercase tracking-wide text-on-surface-variant">{{ label }}</p>
    <p class="mono-numeric mt-2 text-2xl font-semibold text-primary">{{ value }}</p>
    <p v-if="detail" class="mt-2 text-xs text-on-surface-variant">{{ detail }}</p>
  </section>
</template>

<style scoped>
  section {
    position: relative;
    overflow: hidden;
    transition: all 0.15s cubic-bezier(0.2, 0, 0, 1);
  }
</style>
