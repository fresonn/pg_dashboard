import * as React from 'react'
import { Loader2Icon } from 'lucide-react'

export type SpinnerProps = React.ComponentProps<'svg'> & {
  size?: number
}

export function Spinner({ size = 5, ...props }: SpinnerProps) {
  return (
    <Loader2Icon
      role="status"
      aria-label="Loading"
      strokeWidth={3}
      className={`size-${size} animate-spin`}
      {...props}
    />
  )
}
