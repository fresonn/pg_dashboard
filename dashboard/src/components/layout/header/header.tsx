import React, { useEffect, useState } from 'react'
import { cva } from 'cva'
import { Separator } from '@/components/ui/shadcn/separator'
import { SidebarTrigger } from '@/components/ui/shadcn/sidebar'
import { Typography } from '@/components/ui/typography'
import { ConnectionStatus } from './connection-status'

export type HeaderProps = React.HTMLAttributes<HTMLElement> & {
  title: string
  fixed?: boolean
  ref?: React.Ref<HTMLElement>
}

const headerStyle = cva('z-50 mb-10 h-20 rounded-lg', {
  variants: {
    fixed: {
      true: 'header-fixed peer/header sticky top-0 w-[inherit]',
      false: ''
    },
    hasShadow: {
      true: 'shadow',
      false: 'shadow-none'
    }
  },
  compoundVariants: [
    {
      fixed: false,
      hasShadow: true,
      class: 'shadow-none'
    }
  ],
  defaultVariants: {
    fixed: false,
    hasShadow: false
  }
})

const contentStyle = cva('relative flex h-full items-center gap-3 p-4 sm:gap-4', {
  variants: {
    blurred: {
      true: 'after:bg-background/20 after:absolute after:inset-0 after:-z-10 after:backdrop-blur-lg',
      false: ''
    }
  },
  defaultVariants: {
    blurred: false
  }
})

export function Header({ title, className, fixed, children, ...props }: HeaderProps) {
  const [offset, setOffset] = useState(0)

  useEffect(() => {
    const onScroll = () => {
      setOffset(document.body.scrollTop || document.documentElement.scrollTop)
    }

    document.addEventListener('scroll', onScroll, { passive: true })

    return () => document.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <header
      className={headerStyle({
        fixed,
        hasShadow: fixed && offset > 10,
        className
      })}
      {...props}
    >
      <div
        className={contentStyle({
          blurred: fixed && offset > 10
        })}
      >
        <SidebarTrigger variant="outline" className="size-8" />
        <Separator orientation="vertical" className="h-9!" />
        <Typography variant="h2" as="h1">
          {title}
        </Typography>
        {children}
        <div className="ml-auto">
          <ConnectionStatus />
        </div>
      </div>
    </header>
  )
}
