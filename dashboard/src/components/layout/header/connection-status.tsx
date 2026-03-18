import { Separator } from '@/components/ui/shadcn/separator'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import { Typography } from '@/components/ui/typography'
import { useGetStatus } from '@/lib/api/gen'
import { Database, UserRound } from 'lucide-react'

export function ConnectionStatus() {
  const { data } = useGetStatus()

  return (
    <div className="bg-theme-color/10 border-theme-color/50 text-theme-color flex items-center rounded-lg border px-2.5 py-1.5">
      <Tooltip>
        <TooltipTrigger className="mr-2 flex items-center">
          <UserRound className="mr-1" size={20} />
          <Typography variant="small" as="span">
            {data?.user}
          </Typography>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <Typography className="font-medium">Current user</Typography>
        </TooltipContent>
      </Tooltip>
      <Separator orientation="vertical" className="bg-theme-color/60 h-5!" />
      <Tooltip>
        <TooltipTrigger className="ml-2 flex items-center">
          <Database className="mr-1" strokeWidth={1.5} size={20} />
          <Typography variant="small" as="span">
            {data?.database}
          </Typography>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <Typography className="font-medium">Current database</Typography>
        </TooltipContent>
      </Tooltip>
    </div>
  )
}
