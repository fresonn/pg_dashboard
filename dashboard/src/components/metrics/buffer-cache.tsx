import ContentLoader from 'react-content-loader'
import { Typography } from '@/components/ui/typography'
import { MaybeTooltip } from '@/components/ui/maybe-tooltip'
import { cn } from '@/lib/utils'
import {
  HIT_RATE_DANGER_PERCENT,
  HIT_RATE_HEALTHY_PERCENT,
  clampPercent,
  formatRate,
  hitRateTone,
  type BufferMetricScope,
  type HitRateTone
} from '@/lib/metrics'
import { useBufferCacheMetrics } from '@/hooks/use-buffer-cache-metrics'

/** Number of segments of the bar gauge. */
const GAUGE_SEGMENTS = 24

/** Percent step of the scale printed under the gauge (numbers only). */
const GAUGE_SCALE_STEP = 25

const GAUGE_SCALE = Array.from(
  { length: 100 / GAUGE_SCALE_STEP + 1 },
  (_, index) => index * GAUGE_SCALE_STEP
)

/** Colours of the gauge levels, taken from the design system. */
const TONE_CLASS: Record<HitRateTone, { value: string; bar: string }> = {
  bad: { value: 'text-destructive', bar: 'bg-destructive' },
  warning: { value: 'text-yellow-400', bar: 'bg-yellow-400' },
  good: { value: 'text-theme-color', bar: 'bg-theme-color' }
}

export function BufferCacheWidgetSkeleton() {
  return (
    <ContentLoader
      className="h-full"
      backgroundColor="var(--skeleton-bg)"
      foregroundColor="var(--skeleton-fg)"
      width="100%"
    >
      <rect x="0" y="0" width="30%" height={22} rx="6" ry="6" />
      <rect x="0" y={34} width="100%" height={10} rx="3" ry="3" />
      <rect x="0" y={58} width="30%" height={8} rx="4" ry="4" />
      <rect x="0" y={72} width="20%" height={12} rx="4" ry="4" />
      <rect x="0" y={90} width="24%" height={8} rx="4" ry="4" />
      <rect x="52%" y={58} width="30%" height={8} rx="4" ry="4" />
      <rect x="52%" y={72} width="20%" height={12} rx="4" ry="4" />
      <rect x="52%" y={90} width="24%" height={8} rx="4" ry="4" />
    </ContentLoader>
  )
}

/**
 * Live shared buffers cache hit rate.
 *
 * The gauge shows the hit ratio of the requested scope exactly as the API
 * reports it: Postgres already rounds it to two decimals, so the number is
 * printed as is and always matches the SSE payload and GET /api/metrics/history
 * (no client side rounding or reformatting). The hint next to it appears only
 * while the first measurement is on its way; the value itself never disappears,
 * so an idle cluster or a reconnecting stream keeps the last number on screen.
 *
 * `scope="cluster"` reads the cluster wide metric (Cluster Overview),
 * `scope="database"` the metric of the connected database (database overview).
 *
 * The two plain text stats below are the live traffic: hits/s and misses/s of
 * the last collection window, derived from the cumulative counters. Their window
 * and the window hit rate are described in the tooltip.
 */
export function BufferCacheWidget({ scope = 'cluster' }: { scope?: BufferMetricScope }) {
  const { cumulativeHitRate, hitRate, lastKnownHitRate, hitsPerSecond, readsPerSecond, windowMs } =
    useBufferCacheMetrics(scope)

  const rate = cumulativeHitRate
  const windowRate = hitRate ?? lastKnownHitRate

  const hasRate = rate !== null
  const tone = hasRate ? hitRateTone(rate) : null
  const filledSegments = hasRate ? Math.round((clampPercent(rate) / 100) * GAUGE_SEGMENTS) : 0

  const hint = hasRate ? null : 'waiting for samples'

  const windowSeconds = windowMs === null ? null : (windowMs / 1000).toFixed(1)

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-end gap-2">
        <MaybeTooltip
          side="bottom"
          content={
            <p>
              buffer_cache_hit: cumulative ratio calculated by Postgres since the last statistics
              reset, shown exactly as the API reports it. Red below {HIT_RATE_DANGER_PERCENT}%,
              amber from {HIT_RATE_DANGER_PERCENT}% to {HIT_RATE_HEALTHY_PERCENT}%, green from{' '}
              {HIT_RATE_HEALTHY_PERCENT}%.
              {windowRate !== null &&
                ` Window rate${windowSeconds ? ` (last ${windowSeconds}s)` : ''}: ${windowRate.toFixed(1)}%.`}
            </p>
          }
        >
          <Typography
            variant="h3"
            as="span"
            className={cn(
              'font-semibold tabular-nums',
              tone ? TONE_CLASS[tone].value : 'text-muted-foreground'
            )}
          >
            {hasRate ? `${rate}%` : '—'}
          </Typography>
        </MaybeTooltip>

        {hint && (
          <Typography variant="small" className="text-muted-foreground pb-1">
            {hint}
          </Typography>
        )}
      </div>

      <div
        role="meter"
        aria-label="Shared buffers cache hit rate"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={hasRate ? rate : undefined}
        className="mt-2 flex gap-[3px]"
      >
        {Array.from({ length: GAUGE_SEGMENTS }, (_, index) => (
          <span
            key={index}
            className={cn(
              'h-3 flex-1 rounded-[2px] transition-colors duration-500',
              index < filledSegments && tone ? TONE_CLASS[tone].bar : 'bg-white/10'
            )}
          />
        ))}
      </div>

      {/* Percent scale of the bar: numbers only, the sign lives in the tooltip. */}
      <div
        aria-hidden="true"
        className="text-muted-foreground mt-1 flex justify-between text-[10px] leading-none tabular-nums"
      >
        {GAUGE_SCALE.map((value) => (
          <span key={value}>{value}</span>
        ))}
      </div>

      <div className="mt-3 grid grid-cols-2 gap-3">
        <TrafficStat title="Buffer Hits" value={hitsPerSecond} label="Hits/s" />
        <TrafficStat title="Buffer Reads" value={readsPerSecond} label="Misses/s" />
      </div>
    </div>
  )
}

/** Plain text traffic counter: no card, no background. */
function TrafficStat({
  title,
  value,
  label
}: {
  title: string
  value: number | null
  label: string
}) {
  return (
    <div className="flex flex-col">
      <Typography variant="small" className="text-muted-foreground text-xs leading-tight">
        {title}
      </Typography>

      <Typography
        variant="small"
        as="span"
        className="text-lg leading-tight font-medium tabular-nums"
      >
        {formatRate(value)}
        <span className="text-muted-foreground text-xs font-normal">/s</span>
      </Typography>

      <Typography variant="small" className="text-muted-foreground text-xs leading-tight">
        {label}
      </Typography>
    </div>
  )
}
