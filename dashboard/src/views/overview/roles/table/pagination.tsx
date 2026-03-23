import { Button } from '@/components/ui/button'
import { ChevronRightIcon } from 'lucide-react'
import type { Table } from '@tanstack/react-table'

export interface DataTablePaginationProps<TData> {
  table: Table<TData>
}

export function TablePagination<TData>({ table }: DataTablePaginationProps<TData>) {
  return (
    <div className="flex items-center">
      <Button
        type="button"
        variant="outline"
        className="mr-2.5 pr-2 pl-2"
        onClick={() => table.previousPage()}
        disabled={!table.getCanPreviousPage()}
      >
        <ChevronRightIcon className="rotate-180" />
      </Button>
      <Button
        variant="outline"
        className="pr-2 pl-2"
        onClick={() => table.nextPage()}
        disabled={!table.getCanNextPage()}
      >
        <ChevronRightIcon />
      </Button>
    </div>
  )
}
