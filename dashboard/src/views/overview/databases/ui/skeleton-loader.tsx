import ContentLoader from 'react-content-loader'

export function DatabasesTableSkeleton({
  rows,
  isError = false
}: {
  rows: number
  isError?: boolean
}) {
  const rowHeight = 40
  const startY = 16

  return (
    <ContentLoader
      height={rowHeight * rows + 6}
      width="100%"
      backgroundColor={isError ? 'var(--color-red-500)' : 'var(--skeleton-bg)'}
      foregroundColor={isError ? 'var(--color-red-400)' : 'var(--skeleton-fg)'}
    >
      <rect x={0} y={0} width="100%" height="1" />
      {Array.from({ length: 10 }).map((_, i) => (
        <rect
          key={`header-${i}`}
          x={`calc(${i * 10}% + 0.8%)`} // a small indentation on the left inside the column
          y={14}
          rx={6}
          width="calc(10% - 1.6%)" // subtract the left and right indents
          height={18}
        />
      ))}

      {Array.from({ length: 11 }).map((_, i) => (
        <rect
          key={`vline-${i}`}
          x={i === 10 ? '99.9%' : `${i * 10}%`}
          y={0}
          width="1"
          height="100%"
          rx="0.5"
        />
      ))}

      {Array.from({ length: rows }).map((_, i) => {
        const yPos = startY + 29 + i * rowHeight // 38 = headerheight + some gap
        return <rect key={`hline-${i}`} x="0%" y={yPos} width="100%" height="1" />
      })}
    </ContentLoader>
  )
}
