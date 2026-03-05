import { DatabasesTableSkeleton } from './skeleton-loader'

import { InlineErrorBanner } from './error-banner'

export function TableErrorState({ onRefetch }: { onRefetch(): void }) {
  return (
    <div className="relative">
      <DatabasesTableSkeleton isError rows={11} />
      <div className="absolute inset-0 z-10 flex items-center justify-center bg-black/30 backdrop-blur-xs">
        <InlineErrorBanner onRefetch={onRefetch} />
      </div>
    </div>
  )
}
