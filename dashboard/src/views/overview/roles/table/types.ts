import type { RoleAccessLevel, RoleView, RoleAttributes } from '@/lib/api/gen'

export type Role = {
  id: string
  name: string
  isGroup: boolean
  accessLevel: RoleAccessLevel
  membership: RoleView['membership']
  attributes: RoleAttributes[]
}
