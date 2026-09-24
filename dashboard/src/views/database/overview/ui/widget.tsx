import { Suspense, type ReactNode } from 'react'
import { cva } from 'cva'
import { Typography } from '@/components/ui/typography'
import { ErrorBoundary } from '@/components/error-boundary'
import { TriangleAlert } from 'lucide-react'

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

const widgetContainer = cva('bg-secondary shadow-surface rounded-lg border p-5')

export function Widget({
  title,
  children,
  skeleton,
  className,
  icon
}: {
  title: string
  children: ReactNode
  skeleton: ReactNode
  icon: ReactNode
  className?: string
}) {
  return (
    <div className={widgetContainer({ className })}>
      <div className="flex h-full flex-col">
        <div className="mb-4 flex items-center">
          <div className="text-theme-color mr-3.5">{icon}</div>
          <Typography className="font-medium">{title}</Typography>
        </div>
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
