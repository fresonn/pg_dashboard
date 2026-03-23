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

export const roleFlag = {
  superuser: 'superuser',
  login: 'login',
  createRole: 'create_role',
  createDb: 'create_db',
  replication: 'replication'
} as const

type RoleFlagKey = (typeof roleFlag)[keyof typeof roleFlag]

const roleFlagConfig: Record<RoleFlagKey, { icon: LucideIcon; color: string; label: string }> = {
  superuser: { icon: UserStar, color: 'text-sky-400', label: 'Super user' },
  login: { icon: LogIn, color: 'text-teal-400', label: 'Log in' },
  create_role: { icon: UserRoundPlus, color: 'text-yellow-400', label: 'Create role' },
  create_db: { icon: Database, color: 'text-cyan-400', label: 'Create database' },
  replication: { icon: DatabaseBackup, color: 'text-violet-400', label: 'Replication' }
}

export function RoleFlag({ flag, iconSize = 22 }: { flag: string; iconSize?: number }) {
  const config = roleFlagConfig[flag as RoleFlagKey]
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
