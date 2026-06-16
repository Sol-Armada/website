export type RefreshMode = 'background' | 'blocking'

function resolveNextMode(currentMode: RefreshMode | null, isBackground: boolean): RefreshMode {
    return !isBackground || currentMode === 'blocking'
        ? 'blocking'
        : 'background'
}

export function createRequestQueue() {
    let inFlightRequest: Promise<void> | null = null
    let queuedRefreshMode: RefreshMode | null = null

    async function run(
        options: { background?: boolean } = {},
        requestFactory: (isBackground: boolean) => Promise<void>,
    ): Promise<void> {
        const isBackground = options.background === true

        if (inFlightRequest !== null) {
            queuedRefreshMode = resolveNextMode(queuedRefreshMode, isBackground)
            await inFlightRequest
            return
        }

        const request = requestFactory(isBackground)
        inFlightRequest = request
        await request
        inFlightRequest = null

        if (queuedRefreshMode !== null) {
            const nextMode = queuedRefreshMode
            queuedRefreshMode = null
            void run({ background: nextMode === 'background' }, requestFactory)
        }
    }

    function clear() {
        queuedRefreshMode = null
    }

    return {
        run,
        clear,
    }
}
