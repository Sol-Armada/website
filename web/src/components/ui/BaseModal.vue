<script setup lang="ts">
  interface Props {
    isOpen: boolean
    title?: string
    loading?: boolean
    error?: string | null
    maxWidth?: string
  }

  interface Emits {
    (e: 'close'): void
  }

  withDefaults(defineProps<Props>(), {
    title: '',
    loading: false,
    error: null,
    maxWidth: '40rem',
  })

  const emit = defineEmits<Emits>()

  function handleOverlayClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      emit('close')
    }
  }
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="isOpen"
        class="modal-overlay"
        @click="handleOverlayClick"
      >
        <div
          class="modal-panel"
          :style="{ maxWidth }"
        >
          <!-- Header -->
          <div class="modal-header">
            <h3 v-if="title" class="modal-title">{{ title }}</h3>
            <slot name="header" />

            <button
              class="modal-close"
              type="button"
              @click="emit('close')"
            >
              <span aria-hidden="true">×</span>
            </button>
          </div>

          <!-- Error message -->
          <div v-if="error" class="modal-error">
            {{ error }}
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot />
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
  .modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 70;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: rgb(0 0 0 / 0.5);
    backdrop-filter: blur(4px);
  }

  .modal-panel {
    width: 100%;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    background: var(--v0-surface);
    box-shadow: 0 24px 72px rgb(0 0 0 / 0.45);
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.25rem;
    border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
  }

  .modal-title {
    color: var(--sa-gold);
    font-size: 1.125rem;
    font-weight: 700;
  }

  .modal-close {
    background: none;
    border: none;
    color: var(--sa-muted);
    width: 2rem;
    height: 2rem;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 1.5rem;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-close:hover {
    background: rgb(148 163 184 / 0.14);
    color: var(--sa-fg);
  }

  .modal-error {
    background: rgba(251, 113, 133, 0.1);
    border: 1px solid rgba(251, 113, 133, 0.3);
    color: var(--sa-danger);
    padding: 0.75rem 1.25rem;
    font-size: 0.875rem;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    padding: 1.25rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 1rem 1.25rem;
    border-top: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
  }

  /* Transition animations */
  .modal-fade-enter-active,
  .modal-fade-leave-active {
    transition: opacity 0.2s ease;
  }

  .modal-fade-enter-from,
  .modal-fade-leave-to {
    opacity: 0;
  }

  .modal-fade-enter-active .modal-panel,
  .modal-fade-leave-active .modal-panel {
    transition: transform 0.2s ease, opacity 0.2s ease;
  }

  .modal-fade-enter-from .modal-panel,
  .modal-fade-leave-to .modal-panel {
    opacity: 0;
    transform: scale(0.95);
  }
</style>
