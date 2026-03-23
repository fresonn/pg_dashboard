import { Typography } from '@/components/ui/typography'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import {
  Database,
  DatabaseBackup,
  LogIn,
  UserRoundPlus,
  UserStar,
  type LucideIcon
} from 'lucide-react'
import { type RoleAttributes } from '@/lib/api/gen'

const roleAttributeConfig: Record<
  RoleAttributes,
  { icon: LucideIcon; color: string; label: string }
> = {
  superuser: { icon: UserStar, color: 'text-sky-400', label: 'Super user' },
  login: { icon: LogIn, color: 'text-teal-400', label: 'Log in' },
  createRole: { icon: UserRoundPlus, color: 'text-yellow-400', label: 'Create role' },
  createDatabase: { icon: Database, color: 'text-cyan-400', label: 'Create database' },
  replication: { icon: DatabaseBackup, color: 'text-violet-400', label: 'Replication' }
}

export function RoleAttributeFlag({
  attribute,
  iconSize = 22
}: {
  attribute: RoleAttributes
  iconSize?: number
}) {
  const config = roleAttributeConfig[attribute]
  if (!config) return null

  const { icon: Icon, color, label } = config

  return (
    <div>
      <Tooltip>
        <TooltipTrigger className="flex items-center">
          <Icon className={color} size={iconSize} />
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <Typography>{label}</Typography>
        </TooltipContent>
      </Tooltip>
    </div>
  )
}
