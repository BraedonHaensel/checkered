import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * Utility function for merging className inputs.
 * @param inputs className inputs to merge.
 * @returns The merged className inputs.
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

/**
 * Sleep for a given amount of time.
 * @param ms Number of milliseconds to sleep.
 * @returns A promise that resolves after the given amount of time.
 */
export function sleep(ms: number): Promise<void> {
    // SOURCE: https://stackoverflow.com/questions/951021/what-is-the-javascript-version-of-sleep
    return new Promise((resolve) => setTimeout(resolve, ms))
}
