import { useState } from 'react'
import { keepPreviousData } from '@tanstack/react-query'
import type { SortingState } from '@tanstack/react-table'
import { useQueryParams } from './params'
import { DatabasesTable } from './table/table'
import { useDatabasesDetailed } from '@/lib/api/gen'
import { DatabasesTableSkeleton } from './ui/skeleton-loader'
import { TableErrorState } from './ui/databases-table-error'

export function AvailableDatabases() {
  const [sorting, setSorting] = useState<SortingState>([])

  const queryParams = useQueryParams(sorting)

  const { data, isLoading, isFetching, isError, refetch } = useDatabasesDetailed(queryParams, {
    query: {
      placeholderData: keepPreviousData,
      staleTime: 30_000,
      retry: false,
      refetchOnWindowFocus: false
    }
  })

  if (isLoading && !data) {
    return <DatabasesTableSkeleton rows={11} />
  }

  if (isError && !data) {
    return <TableErrorState onRefetch={refetch} />
  }

  return (
    <DatabasesTable
      data={data ?? []}
      isFetching={isFetching}
      sorting={sorting}
      onSortingChange={setSorting}
    />
  )
}
