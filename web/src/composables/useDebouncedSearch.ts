import { onUnmounted, ref, watch } from 'vue'

export interface UseDebouncedSearchOptions {
  delay?: number
  onSearch: (value: string) => void | Promise<void>
}

/**
 * Composable for debounced search input
 * Automatically cleans up timer on unmount
 * 
 * @param options - Configuration options
 * @returns Reactive search value and clear function
 * 
 * @example
 * const { search, clearDebounce } = useDebouncedSearch({
 *   delay: 300,
 *   onSearch: async (value) => {
 *     await loadResults(value)
 *   }
 * })
 */
export function useDebouncedSearch(options: UseDebouncedSearchOptions) {
  const { delay = 300, onSearch } = options
  const search = ref('')
  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  function clearDebounce() {
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
  }

  watch(search, (value) => {
    clearDebounce()
    debounceTimer = setTimeout(() => {
      void onSearch(value)
    }, delay)
  })

  onUnmounted(() => {
    clearDebounce()
  })

  return {
    search,
    clearDebounce,
  }
}
