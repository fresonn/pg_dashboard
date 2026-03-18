import { createFileRoute } from '@tanstack/react-router'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { clusterStatusQuery } from '@/lib/api/cluster-status'
import { OfflineApiGuard } from '@/components/offline-api-guard'
import { useEffect, useState } from 'react'

export const Route = createFileRoute('/_authenticated')({
  component: () => {
    const [retrying, setRetrying] = useState(false)
    const { clusterStatus } = Route.useRouteContext()

    const queryClient = useQueryClient()

    const { isError, isFetching } = useQuery({
      ...clusterStatusQuery,
      placeholderData: clusterStatus,
      staleTime: Infinity,
      gcTime: Infinity,
      refetchOnWindowFocus: false,
      refetchOnReconnect: false,
      refetchOnMount: false,
      retry: false
    })

    useEffect(() => {
      if (!isFetching) setRetrying(false)
    }, [isFetching])

    const handleRetry = () => {
      setRetrying(true)
      queryClient.refetchQueries({ queryKey: clusterStatusQuery.queryKey })
    }

    return (
      <>
        <OfflineApiGuard open={isError || retrying} onRetry={handleRetry} loading={isFetching} />
        <AuthenticatedLayout />
      </>
    )
  }
})
