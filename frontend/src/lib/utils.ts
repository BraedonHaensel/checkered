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
