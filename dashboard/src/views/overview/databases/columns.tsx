import type { Database } from './types'

import { createColumnHelper } from '@tanstack/react-table'
import { BooleanCheckState } from '@/components/ui/boolean-check-state'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/shadcn/tooltip'
import { Typography } from '@/components/ui/typography'

const columnHelper = createColumnHelper<Database>()

export const columns = [
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('owner', {
    header: 'Owner',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('sizePretty', {
    header: 'Size',
    cell: (props) => {
      const db = props.row.original

      return (
        <Tooltip>
          <TooltipTrigger>{props.getValue()}</TooltipTrigger>
          <TooltipContent side="bottom">
            <Typography>
              <Typography as="span" className="font-medium">
                Raw byte size:
              </Typography>
              <Typography as="span" className="ml-1">
                {db.sizeBytes}
              </Typography>
            </Typography>
          </TooltipContent>
        </Tooltip>
      )
    }
  }),
  columnHelper.accessor('connectionLimit', {
    header: 'Connection Limit',
    // -1 — unlimited
    // 0 — connections is not allowed
    // >0 — max connections
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('encoding', {
    header: 'Encoding',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('ctype', {
    header: 'C-type',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('collate', {
    header: 'Collate',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('activeConnections', {
    header: 'Active Connections',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('isTemplate', {
    header: 'Template',
    cell: (props) => <BooleanCheckState check={props.getValue()} />
  }),
  columnHelper.accessor('allowConnections', {
    header: 'Allow Connections',
    cell: (props) => <BooleanCheckState check={props.getValue()} />
  })
]
