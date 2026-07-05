<script lang="ts" setup>
  //
</script>

<template>
  <v-app>
    <div aria-hidden="true" class="hex-grid-bg" />

    <v-main>
      <main class="main-bg min-h-screen">
        <router-view />
      </main>
    </v-main>
  </v-app>
</template>

<style>
  @import './assets/sol-armada-tactical.css';

  .v-application {
    background: transparent !important;
  }

  .v-main {
    padding: 0 !important;
  }

  .main-bg {
    background: color-mix(in srgb, var(--v0-background) 85%, transparent);
  }
</style>
