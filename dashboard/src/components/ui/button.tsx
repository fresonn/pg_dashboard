import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'cva'
import { Spinner } from './spinner'

export type ButtonProps = React.ComponentProps<'button'> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
    loading?: boolean
  }

export const buttonVariants = cva(
  "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/40 aria-invalid:border-destructive relative inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md text-base font-medium whitespace-nowrap transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default: 'surface-raised',
        destructive: 'focus-visible:ring-destructive/40 bg-red-500 text-white hover:bg-red-400',
        outline:
          'hover:text-accent-foreground bg-input/30 border-input hover:bg-input/50 border shadow-xs',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
        light: 'bg-gray-50 text-black hover:bg-gray-200',
        ghost: 'hover:text-accent-foreground hover:bg-accent/50',
        link: 'text-primary underline-offset-4 hover:underline'
      },
      size: {
        default: 'h-9 px-4 py-2 has-[>svg]:px-3',
        sm: 'h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5',
        lg: 'h-10 rounded-md px-6 has-[>svg]:px-4',
        icon: 'size-9',
        'icon-sm': 'size-8',
        'icon-lg': 'size-10'
      },
      fullWidth: {
        true: 'w-full',
        false: 'w-auto'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'default'
    }
  }
)

export function Button({
  className,
  variant,
  size,
  asChild = false,
  children,
  loading,
  onClick,
  fullWidth = false,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : 'button'

  const onClickMiddleware: React.MouseEventHandler<HTMLButtonElement> = (e) => {
    if (loading) {
      e.preventDefault()
      return
    }

    onClick?.(e)
  }

  return (
    <Comp
      data-slot="button"
      className={buttonVariants({ fullWidth, variant, size, className })}
      onClick={onClickMiddleware}
      {...props}
    >
      {loading && (
        <div className="absolute inset-0 flex items-center justify-center text-current">
          <Spinner />
        </div>
      )}
      <span
        className={`inline-flex items-center justify-center gap-2 ${
          loading
            ? 'invisible opacity-0 transition-opacity'
            : 'visible opacity-100 transition-opacity'
        }`}
      >
        {children}
      </span>
    </Comp>
  )
}
