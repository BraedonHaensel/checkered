'use client'

import { useEffect, useRef, useState, type ReactNode } from 'react'
import { cn } from '../lib/utils'

type Props = {
    className?: string
    children: ReactNode
}

/**
 * Square container that takes the maximum available size.
 */
export default function MaxSquare({ className = '', children }: Props) {
    const [size, setSize] = useState(0)
    const measurementDivRef = useRef<HTMLDivElement>(null)

    // Resize the square when the parent size changes
    useEffect(() => {
        const measurementDiv = measurementDivRef.current
        if (!measurementDiv) return

        const observer = new ResizeObserver(() => {
            setSize(
                Math.min(
                    measurementDiv.clientWidth,
                    measurementDiv.clientHeight
                )
            )
        })
        observer.observe(measurementDiv)
        return () => observer.disconnect()
    })

    return (
        <div className="relative h-full w-full overflow-hidden">
            {/* Empty div used to measure the available width and height */}
            <div ref={measurementDivRef} className="absolute inset-0" />
            {/* Content container dynamically set to the maximum square size */}
            <div
                className={cn(
                    // Center within the available space (if rectangular)
                    'absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
                    className
                )}
                style={{ width: size, height: size }}
            >
                {children}
            </div>
        </div>
    )
}
