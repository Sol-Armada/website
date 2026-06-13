<script lang="ts" setup>
  //
</script>

<template>
  <div aria-hidden="true" class="hex-grid-bg" />

  <main class="main-bg min-h-screen">
    <router-view />
  </main>
</template>

<style>
  @import './assets/sol-armada-tactical.css';

  .main-bg {
    background: color-mix(in srgb, var(--v0-background) 85%, transparent);
  }
</style>
