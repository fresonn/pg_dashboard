import { Ellipsis, UserRoundX } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Typography } from '@/components/ui/typography'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger
} from '@/components/ui/shadcn/dropdown-menu'

export interface MembershipListProps {
  membership: Array<{ name: string; description: string }>
}

export function MembershipList({ membership }: MembershipListProps) {
  if (membership === null) {
    return (
      <div>
        <Tooltip>
          <TooltipTrigger className="mr-2 flex items-center">
            <UserRoundX className="text-muted-foreground h-4" />
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <Typography>Has no membership</Typography>
          </TooltipContent>
        </Tooltip>
      </div>
    )
  }

  return (
    <div className="flex items-center">
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild className="h-6">
          <Button variant="ghost" size="sm" className="data-[state=open]:bg-muted t size-5!">
            <Ellipsis />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="center">
          <ul className="p-2">
            {membership.map((m, ind) => (
              <li key={ind}>
                <Typography as="span" variant="small">
                  {m.name}
                </Typography>
              </li>
            ))}
          </ul>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
