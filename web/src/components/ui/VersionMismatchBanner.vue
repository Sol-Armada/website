<script setup lang="ts">
  import { Snackbar } from '@vuetify/v0'
  import { computed, ref, watch } from 'vue'

  interface Props {
    portalVersion: string
    serverVersion: string
  }

  const props = defineProps<Props>()
  const snackbarOpen = ref(false)
  const snackbarId = 'version-mismatch'

  const isVersionMismatch = computed(() => {
    const detectedVersion = props.serverVersion.trim()
    if (!detectedVersion) {
      return false
    }
    return detectedVersion !== props.portalVersion
  })

  const versionMessage = computed(() => {
    return `A new version is available (${props.serverVersion}). Refresh to update.`
  })

  function refreshForLatestVersion() {
    window.location.reload()
  }

  // Show snackbar when version mismatch is detected
  watch(
    isVersionMismatch,
    newVal => {
      if (newVal) {
        snackbarOpen.value = true
      }
    },
    { immediate: true },
  )
</script>

<template>
  <Snackbar.Portal>
    <Snackbar.Root v-if="snackbarOpen" :id="snackbarId" :key="snackbarId" class="fixed bottom-4 right-4 z-50 max-w-md rounded-lg bg-primary/95 px-6 py-4 shadow-lg">
      <div class="flex items-center justify-between gap-4 text-white">
        <Snackbar.Content>{{ versionMessage }}</Snackbar.Content>

        <div class="flex items-center gap-2">
          <button
            class="rounded px-3 py-1 text-sm font-semibold bg-white/20 hover:bg-white/30 transition-colors"
            type="button"
            @click="refreshForLatestVersion"
          >
            Refresh
          </button>

          <Snackbar.Close class="text-white hover:text-white/80" />
        </div>
      </div>
    </Snackbar.Root>
  </Snackbar.Portal>
</template>
