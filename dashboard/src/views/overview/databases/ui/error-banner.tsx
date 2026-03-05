import { Button } from '@/components/ui/button'
import { Typography } from '@/components/ui/typography'
import { RotateCcw, TriangleAlert } from 'lucide-react'

export function InlineErrorBanner({ onRefetch }: { onRefetch(): void }) {
  return (
    <div className="pt-5">
      <div className="border-destructive/50 bg-destructive/10 text-destructive mb-2 flex items-center rounded-lg border px-4 py-4 text-sm">
        <div className="mr-2 text-red-400">
          <TriangleAlert size={60} strokeWidth={1.5} />
        </div>
        <div>
          <Typography className="leading-none font-medium">Failed to retrieve data</Typography>
          <Typography as="span" className="text-destructive/80">
            Try to find out what is happening in the logs.
          </Typography>
        </div>
      </div>
      <div className="flex justify-center">
        <Button variant="outline" size="sm" onClick={onRefetch}>
          <RotateCcw />
        </Button>
      </div>
    </div>
  )
}
