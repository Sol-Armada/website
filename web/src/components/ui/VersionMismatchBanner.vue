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
    <Snackbar.Root v-if="snackbarOpen" :id="snackbarId" :key="snackbarId" class="fixed bottom-4 right-4 z-50 max-w-md rounded-lg px-6 py-4 shadow-lg" style="background: rgba(230, 168, 45, 0.95);">
      <div class="flex items-center justify-between gap-4" style="color: var(--sa-bg);">
        <Snackbar.Content>{{ versionMessage }}</Snackbar.Content>

        <div class="flex items-center gap-2">
          <button
            class="rounded px-3 py-1 text-sm font-semibold transition-colors"
            style="background: rgba(9, 11, 18, 0.2); color: var(--sa-bg);"
            type="button"
            @click="refreshForLatestVersion"
          >
            Refresh
          </button>

          <Snackbar.Close style="color: var(--sa-bg); opacity: 0.9;" />
        </div>
      </div>
    </Snackbar.Root>
  </Snackbar.Portal>
</template>
