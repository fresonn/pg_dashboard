import { useState } from 'react'
import { keepPreviousData } from '@tanstack/react-query'
import type { SortingState } from '@tanstack/react-table'
import { useQueryParams } from './params'
import { DatabasesTable } from './table/table'
import { useDatabasesDetailed } from '@/lib/api/gen'
import { DatabasesTableSkeleton } from './table/skeleton-loader'

export function AvailableDatabases() {
  const [sorting, setSorting] = useState<SortingState>([])

  const queryParams = useQueryParams(sorting)

  const { data, isLoading, isFetching, error } = useDatabasesDetailed(queryParams, {
    query: {
      placeholderData: keepPreviousData,
      staleTime: 60_000
    }
  })
  if (isLoading) {
    return <DatabasesTableSkeleton rows={11} />
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
