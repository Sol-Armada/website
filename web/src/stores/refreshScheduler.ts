export function createRefreshScheduler(delayMs = 400) {
    let refreshTimer: number | null = null

    function schedule(task: () => void): void {
        if (refreshTimer !== null) {
            window.clearTimeout(refreshTimer)
        }

        refreshTimer = window.setTimeout(() => {
            refreshTimer = null
            task()
        }, delayMs)
    }

    function clear(): void {
        if (refreshTimer === null) {
            return
        }

        window.clearTimeout(refreshTimer)
        refreshTimer = null
    }

    return {
        schedule,
        clear,
    }
}
