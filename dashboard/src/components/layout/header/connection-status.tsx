import { Separator } from '@/components/ui/shadcn/separator'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import { Typography } from '@/components/ui/typography'
import { useGetStatus } from '@/lib/api/gen'
import { Database, UserRound } from 'lucide-react'

export function ConnectionStatus() {
  const { data } = useGetStatus()

  return (
    <div className="surface-raised flex items-center rounded-xl px-3 py-1.5">
      <Tooltip>
        <TooltipTrigger className="mr-3 flex items-center">
          <UserRound className="mr-2" strokeWidth={2.5} size={18} />
          <Typography variant="small" as="span">
            {data?.user}
          </Typography>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <Typography>Current user</Typography>
        </TooltipContent>
      </Tooltip>
      <Separator orientation="vertical" className="h-4! bg-white/50" />
      <Tooltip>
        <TooltipTrigger className="ml-3 flex items-center">
          <Database className="mr-2" strokeWidth={2} size={18} />
          <Typography variant="small" as="span">
            {data?.database}
          </Typography>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <Typography>Current database</Typography>
        </TooltipContent>
      </Tooltip>
    </div>
  )
}
