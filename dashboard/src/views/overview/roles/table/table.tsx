import { useState } from 'react'
import { columns } from './columns'
import type { Role } from './types'
import { TablePagination } from './pagination'
import {
  getCoreRowModel,
  useReactTable,
  flexRender,
  getFilteredRowModel,
  getFacetedMinMaxValues,
  getPaginationRowModel
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/shadcn/table'

export type RolesTableProps = {
  data: Role[]
}

export function RolesTable({ data }: RolesTableProps) {
  const isFetching = false

  const [pagination, setPagination] = useState({
    pageIndex: 0,
    pageSize: 3
  })

  const table = useReactTable({
    columns,
    data,
    state: {
      pagination
    },
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onPaginationChange: setPagination,
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedMinMaxValues: getFacetedMinMaxValues()
  })

  return (
    <div>
      <div className="relative">
        <Table>
          <TableHeader className="pointer-events-none bg-neutral-900">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} className="text-foreground">
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, {
                          ...header.getContext(),
                          isFetching
                        })}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody loading={isFetching}>
            {table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                data-state={row.getIsSelected() && 'selected'}
                className="cursor-pointer"
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <div className="absolute -top-10 right-0">
          <TablePagination table={table} />
        </div>
      </div>
    </div>
  )
}
