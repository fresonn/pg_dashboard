import { cva } from 'cva'
import { TriangleAlert } from 'lucide-react'
import { Suspense, type ReactNode } from 'react'
import { ErrorBoundary } from '@/components/error-boundary'
import { Typography } from '@/components/ui/typography'

export function WidgetErrorFallback() {
  return (
    <div className="flex h-full flex-col items-center justify-center text-yellow-400">
      <div>
        <TriangleAlert className="size-8" strokeWidth={1.5} />
      </div>
      <Typography variant="small">Failed to fetch</Typography>
    </div>
  )
}

const widgetContainer = cva('bg-secondary shadow-surface rounded-xl p-3')

export function Widget({
  title,
  children,
  skeleton,
  className
}: {
  title: string
  children: ReactNode
  skeleton: ReactNode
  className?: string
  withBackground?: boolean
}) {
  return (
    <div className={widgetContainer({ className })}>
      <div className="flex h-full flex-col">
        <Typography variant="h5" className="text-theme-color mb-3">
          {title}
        </Typography>
        <div className="min-h-0 flex-1">
          <ErrorBoundary fallback={<WidgetErrorFallback />}>
            <Suspense fallback={skeleton}>
              <div className="animate-in fade-in slide-in-from-top-8 duration-300">{children}</div>
            </Suspense>
          </ErrorBoundary>
        </div>
      </div>
    </div>
  )
}
