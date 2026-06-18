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

<style scoped>
  section {
    position: relative;
    overflow: hidden;
  }

  section::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, #E6A82D 0%, #0041FF 50%, transparent 100%);
    opacity: 0.8;
    border-radius: 16px 16px 0 0;
  }
</style>
