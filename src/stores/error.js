import { defineStore } from "pinia"

export const useErrorStore = defineStore("error", () => {
    const error = ref("")
    const show = ref(false)
    const loading = ref(false)
    const timeout = ref(4000)
    const closeable = ref(false)
    const reason = ref("")

    function reset() {
        error.value = ""
        show.value = false
        loading.value = false
        timeout.value = 4000
        closeable.value = false
        reason.value = ""
    }

    return {
        error,
        show,
        loading,
        timeout,
        closeable,
        reason,
        reset,
    }
});
