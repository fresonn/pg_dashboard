import type { Column } from '@tanstack/react-table'
import { ChevronsUpDown, ChevronUp } from 'lucide-react'
import { Spinner } from '../spinner'

type SortableHeaderProps<TData> = {
  title: string
  isFetching: boolean
  column: Column<TData, unknown>
}

export function SortableHeader<TData>({ column, title, isFetching }: SortableHeaderProps<TData>) {
  const isSorted = column.getIsSorted()
  const showLoader = isFetching && isSorted

  return (
    <div
      onClick={column.getToggleSortingHandler()}
      className="flex cursor-pointer items-center gap-0.5 select-none"
    >
      {title}

      {showLoader ? (
        <div className="text-theme-color">
          <Spinner size={4} />
        </div>
      ) : (
        <>
          {isSorted === 'asc' && <ChevronUp className="size-5" />}
          {isSorted === 'desc' && <ChevronUp className="size-5 rotate-180" />}
          {!isSorted && <ChevronsUpDown strokeWidth={3} className="size-4 opacity-50" />}
        </>
      )}
    </div>
  )
}
