<script setup lang="ts">
  import { ref, watch } from 'vue'

  interface Props {
    modelValue: number
    hasNextPage: boolean
    loading?: boolean
  }

  interface Emits {
    (e: 'update:modelValue', value: number): void
    (e: 'first'): void
    (e: 'previous'): void
    (e: 'next'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  })

  const emit = defineEmits<Emits>()

  const pageInput = ref(String(props.modelValue))

  watch(() => props.modelValue, (newPage) => {
    pageInput.value = String(newPage)
  })

  function goToFirstPage() {
    if (props.loading || props.modelValue === 1) return
    emit('first')
    emit('update:modelValue', 1)
  }

  function goToPreviousPage() {
    if (props.loading || props.modelValue <= 1) return
    emit('previous')
    emit('update:modelValue', props.modelValue - 1)
  }

  function jumpToPage() {
    const nextPage = Number.parseInt(pageInput.value, 10)
    if (!Number.isFinite(nextPage) || nextPage < 1) {
      pageInput.value = String(props.modelValue)
      return
    }

    if (nextPage === props.modelValue) {
      return
    }

    emit('update:modelValue', nextPage)
  }

  function goToNextPage() {
    if (!props.hasNextPage || props.loading) return
    emit('next')
    emit('update:modelValue', props.modelValue + 1)
  }
</script>

<template>
  <div class="flex items-center gap-2 text-sm">
    <button
      class="btn-secondary"
      :disabled="modelValue === 1 || loading"
      type="button"
      @click="goToFirstPage"
    >
      First
    </button>

    <button
      class="btn-secondary"
      :disabled="modelValue <= 1 || loading"
      type="button"
      @click="goToPreviousPage"
    >
      Previous
    </button>

    <div class="flex items-center gap-2">
      <span class="text-muted">Page</span>

      <input
        v-model="pageInput"
        class="input-field w-20 text-center"
        :disabled="loading"
        min="1"
        type="number"
        @blur="jumpToPage"
        @keydown.enter="jumpToPage"
      >
    </div>

    <button
      class="btn-secondary"
      :disabled="!hasNextPage || loading"
      type="button"
      @click="goToNextPage"
    >
      Next
    </button>
  </div>
</template>
