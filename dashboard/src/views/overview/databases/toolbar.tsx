import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { X } from 'lucide-react'
import { Typography } from '@/components/ui/typography'

export interface TableToolbarProps<TData> {
  table: Table<TData>
}

export function TableToolbar<TData>({ table }: TableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0

  const totalRows = table.getFilteredRowModel().rows.length

  return (
    <div className="mb-2 flex items-center justify-between">
      <div className="flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2">
        <Input
          placeholder="Database name..."
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(event) => table.getColumn('name')?.setFilterValue(event.target.value)}
          className="h-8 w-80"
        />
        {isFiltered && (
          <Button variant="outline" onClick={() => table.resetColumnFilters()} className="h-8 px-2">
            Discard
            <X className="size-4" />
          </Button>
        )}
      </div>
      <Typography variant="small" className="font-medium">
        Amount: {totalRows}
      </Typography>
    </div>
  )
}
