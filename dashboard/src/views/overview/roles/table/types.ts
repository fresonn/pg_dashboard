import type { RoleAccessLevel, RoleView } from '@/lib/api/gen'

export type Role = {
  id: string
  name: string
  isGroup: boolean
  accessLevel: RoleAccessLevel
  membership: RoleView['membership']
  flags: string[]
  capabilities: string[]
}
