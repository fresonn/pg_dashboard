import { cva } from 'cva'
import { Typography } from '@/components/ui/typography'
import { useSidebar } from '@/components/ui/shadcn/sidebar'

const head = cva('flex items-center transition-all duration-200', {
  variants: {
    open: {
      true: 'mb-4 py-4',
      false: 'mb-2 py-4'
    }
  }
})

const logo = cva('shrink-0 transition-all duration-200', {
  variants: {
    open: {
      true: 'size-10',
      false: 'size-8'
    }
  }
})

const title = cva('pl-1.5 whitespace-nowrap transition-opacity duration-200 select-none', {
  variants: {
    open: {
      true: 'opacity-100',
      false: 'pointer-events-none opacity-0'
    }
  }
})

export function Header() {
  const { open } = useSidebar()

  return (
    <div className={head({ open })}>
      <div className={logo({ open })}>
        <img src="/logo.svg" alt="App logo" />
      </div>
      <Typography className={title({ open })} variant="h4" as="h4">
        PgDashboard
      </Typography>
    </div>
  )
}
