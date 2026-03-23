import type { Role } from './types'
import { RoleFlag } from '../ui/flag'
import { createColumnHelper } from '@tanstack/react-table'
import { MembershipList } from './membership-list'
import { Badge } from '@/components/ui/shadcn/badge'
import { capitalize } from '@/lib/utils'

const columnHelper = createColumnHelper<Role>()

export const columns = [
  columnHelper.accessor('name', {
    header: 'Role',
    cell: (props) => props.getValue()
  }),
  columnHelper.accessor('accessLevel', {
    header: 'Access Level',
    cell: (props) => capitalize(props.getValue())
  }),
  columnHelper.accessor('flags', {
    header: 'Flags',
    cell: ({ row, getValue }) => {
      const flags = getValue()
      const { isGroup } = row.original

      if (flags === null) {
        if (isGroup)
          return (
            <Badge variant="secondary" className="pointer-events-none py-0!">
              Group Role
            </Badge>
          )
      }

      return (
        <ul className="flex items-center">
          {isGroup && (
            <li className="mr-1">
              <Badge variant="secondary" className="pointer-events-none py-0!">
                Group
              </Badge>
            </li>
          )}
          {flags.map((flag, ind) => (
            <li key={ind} className="mr-1">
              <RoleFlag flag={flag} iconSize={20} />
            </li>
          ))}
        </ul>
      )
    }
  }),
  columnHelper.accessor('membership', {
    header: 'Membership',
    cell: (props) => {
      return <MembershipList membership={props.getValue()} />
    }
  })
]
