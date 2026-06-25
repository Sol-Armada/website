<script setup lang="ts">
  defineProps<{
    label: string
    value: string | number
    detail?: string
  }>()
</script>

<template>
  <section class="tactical-panel sweep-animation p-4">
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

  /* section::before {
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

  section:hover {
    border-color: #E6A82D;
    box-shadow: 0 0 24px rgba(230, 168, 45, 0.15);
  }

  section:hover::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(180deg, transparent, rgba(0, 65, 255, 0.1));
    animation: scan-line 0.4s cubic-bezier(0.2, 0, 0, 1);
    pointer-events: none;
  }

  @keyframes scan-line {
    from { transform: translateY(-100%); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  } */
</style>
