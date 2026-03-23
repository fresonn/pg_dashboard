import ContentLoader from 'react-content-loader'
import { useRolesSuspense } from '@/lib/api/gen'
import { RolesTable } from './table/table'

export function RolesWidgetSkeleton() {
  const axisY = [11, 44, 77, 110]

  return (
    <ContentLoader
      className="h-full"
      backgroundColor="var(--skeleton-bg)"
      foregroundColor="var(--skeleton-fg)"
      width="100%"
    >
      {axisY.map((y) => {
        return Array.from({ length: 4 }).map((_, ind) => (
          <rect
            key={`header-${ind}`}
            x={`calc(${ind * 25}% + 0.8%)`}
            y={y}
            rx={6}
            width="calc(25% - 1.6%)"
            height={23}
          />
        ))
      })}
    </ContentLoader>
  )
}

export function RolesWidget() {
  const { data } = useRolesSuspense({
    query: {
      retry: false
    }
  })

  return <RolesTable data={data} />
}
