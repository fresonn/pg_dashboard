import { lazy, Suspense } from 'react'
import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import { type QueryClient } from '@tanstack/react-query'
import { clusterStatusQuery } from '@/lib/api/cluster-status'
import { Toaster } from 'sonner'

export interface RouterContext {
  queryClient: QueryClient
}

const TanStackRouterDevtools = import.meta.env.PROD
  ? () => null
  : lazy(() =>
      import('@tanstack/react-router-devtools').then((res) => ({
        default: res.TanStackRouterDevtools
      }))
    )

export const Route = createRootRouteWithContext<RouterContext>()({
  async beforeLoad({ context }) {
    await context.queryClient.prefetchQuery(clusterStatusQuery)

    const clusterStatus = context.queryClient.getQueryData(clusterStatusQuery.queryKey)

    return {
      clusterStatus: clusterStatus
    }
  },
  component: () => {
    return (
      <>
        <Outlet />
        <Toaster
          position="bottom-right"
          visibleToasts={4}
          toastOptions={{
            classNames: {
              toast: '!border-border !bg-neutral-900 !text-foreground',
              description: '!text-muted-foreground',
              success: '!border-green-500/30 !text-green-500',
              error: '!border-red-500/30 !text-red-400'
            }
          }}
        />
        <Suspense>
          <TanStackRouterDevtools position="bottom-right" />
        </Suspense>
      </>
    )
  }
})
