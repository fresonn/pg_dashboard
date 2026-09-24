import { Typography } from '@/components/ui/typography'
import { Database, ScrollText } from 'lucide-react'
import { HealthyBadge } from './healthy-badge'
import { HoverCard, HoverCardContent, HoverCardTrigger } from '@/components/ui/shadcn/hover-card'

export function DatabaseHeader({
  name,
  owner,
  encoding,
  ctype,
  tablespace,
  isTemplate,
  description
}: {
  name: string
  owner: string
  encoding: string
  ctype: string
  tablespace: string
  isTemplate: boolean
  description?: string
}) {
  const properties = [
    {
      name: 'Owner',
      value: owner
    },
    {
      name: 'Encoding',
      value: encoding
    },
    {
      name: 'C-type',
      value: ctype
    },
    {
      name: 'Tablespace',
      value: tablespace
    },
    {
      name: 'Template',
      value: isTemplate.toString()
    }
  ]

  return (
    <div className="mb-5 flex">
      <div className="bg-secondary surface-raised shadow-surface mr-4 rounded-xl p-5">
        <Database size={30} strokeWidth={1.5} />
      </div>
      <div className="pt-1">
        <div className="mb-1.5 flex items-center">
          <Typography variant="h3" as="h3">
            {name}
          </Typography>
          <div className="ml-4">
            <HealthyBadge />
          </div>
          {description && (
            <HoverCard openDelay={100}>
              <HoverCardTrigger asChild>
                <div className="pl-3">
                  <ScrollText strokeWidth={1.5} size={21} className="text-neutral-300" />
                </div>
              </HoverCardTrigger>
              <HoverCardContent>
                <Typography className="mb-1 font-medium">Description</Typography>
                <Typography variant="small">{description}</Typography>
              </HoverCardContent>
            </HoverCard>
          )}
        </div>
        <ul className="flex items-center">
          {properties.map(({ name, value }, ind) => (
            <li
              key={ind}
              className="0 relative mr-2.5 pr-2.5 after:absolute after:top-1/2 after:right-0 after:h-3.5 after:w-0.5 after:-translate-y-1/2 after:bg-neutral-600 after:content-[''] last:after:hidden"
            >
              <Typography variant="small" className="text-gray-300">
                {name}:
                <Typography as="span" className="ml-1 font-medium">
                  {value}
                </Typography>
              </Typography>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
