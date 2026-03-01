import { usePostgresVersionSuspense } from '@/lib/api/gen'
import { Typography } from '@/components/ui/typography'
import ContentLoader from 'react-content-loader'

export function VersionWidgetSkeleton() {
  return (
    <ContentLoader
      className="h-full"
      backgroundColor="var(--skeleton-bg)"
      foregroundColor="var(--skeleton-fg)"
      width="100%"
    >
      <rect x="0" y="0" width="100%" rx="6" ry="6" height="15" />
      <rect x="0" y="58%" width="45%" rx="6" ry="6" height="15" />
    </ContentLoader>
  )
}

export function VersionWidget() {
  const { data } = usePostgresVersionSuspense({
    query: {
      retry: false
    }
  })

  return (
    <div>
      <Typography variant="code">
        {data.version}
        {data.bitDepth && `, ${data.bitDepth}`}
      </Typography>
    </div>
  )
}
