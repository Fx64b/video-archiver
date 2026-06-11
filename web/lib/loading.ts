/**
 * Waits until at least `minMs` have passed since `start` (a Date.now()
 * timestamp). Used to keep skeleton loaders on screen long enough to avoid a
 * jarring flash when the backend responds quickly.
 */
export async function ensureMinimumDelay(
    start: number,
    minMs = 500
): Promise<void> {
    const remaining = minMs - (Date.now() - start)
    if (remaining > 0) {
        await new Promise((resolve) => setTimeout(resolve, remaining))
    }
}
