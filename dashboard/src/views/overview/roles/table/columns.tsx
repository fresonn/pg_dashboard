import type { Role } from './types'
import { RoleAttributeFlag } from '../ui/attribute-flag'
import { createColumnHelper } from '@tanstack/react-table'
import { MembershipList } from '../ui/membership-list'
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
  columnHelper.accessor('attributes', {
    header: 'Attributes',
    cell: ({ row, getValue }) => {
      const attributes = getValue()
      const { isGroup } = row.original

      if (attributes === null) {
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
          {attributes.map((attr, ind) => (
            <li key={ind} className="mr-1">
              <RoleAttributeFlag attribute={attr} iconSize={20} />
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
