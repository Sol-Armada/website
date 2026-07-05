<script setup lang="ts">
  interface Props {
    modelValue: string
    isRefreshing?: boolean
    loading?: boolean
    placeholder?: string
  }

  interface Emits {
    (e: 'update:modelValue', value: string): void
  }

  withDefaults(defineProps<Props>(), {
    isRefreshing: false,
    loading: false,
    placeholder: 'Search...',
  })

  const emit = defineEmits<Emits>()

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement
    emit('update:modelValue', target.value)
  }
</script>

<template>
  <div>
    <input
      class="input-field w-full"
      :placeholder="placeholder"
      type="search"
      :value="modelValue"
      @input="handleInput"
    >

    <div class="mb-3 mt-2 h-0.5 w-full overflow-hidden rounded-full bg-surface-variant/40">
      <div
        class="h-full w-full bg-primary/80 transition-opacity duration-150"
        :class="isRefreshing && !loading ? 'animate-pulse opacity-100' : 'opacity-0'"
      />
    </div>
  </div>
</template>
