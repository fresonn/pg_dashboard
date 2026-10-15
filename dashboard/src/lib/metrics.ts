/**
 * Shared buffers metrics: the wire contract of the SSE stream plus the pure
 * helpers the live widgets are built on.
 *
 * The stream endpoint is registered next to the OpenAPI router (a streaming
 * response cannot be expressed in the strict server interface), so its payload
 * shape is declared here and has to stay in sync with `internal/model/metrics`
 * on the backend:
 *
 *   event: metrics
 *   data: {"buffer_cache_hit":{"value":99.86,"timestamp":"2026-10-11T22:00:59.792069+03:00"}}
 */
/**
 * Scope of the shared buffers metrics.
 *
 * The backend produces both scopes with a single pg_stat_database query:
 * `cluster` aggregates every database, `database` only the database the
 * application is connected to (the one the database overview page shows).
 */
export type BufferMetricScope = 'cluster' | 'database'

export type BufferMetricSet = {
  /** Cumulative hit ratio calculated by Postgres since the last statistics reset */
  cumulativeHitRate: string
  /** Cumulative number of shared buffers hits */
  hitsTotal: string
  /** Cumulative number of shared buffers reads, i.e. cache misses */
  readsTotal: string
}

/**
 * Metric families of the shared buffers pipeline, mirroring
 * `internal/model/metrics` on the backend.
 */
export const BUFFER_METRICS: Record<BufferMetricScope, BufferMetricSet> = {
  cluster: {
    cumulativeHitRate: 'buffer_cache_hit',
    hitsTotal: 'buffer_hits_total',
    readsTotal: 'buffer_reads_total'
  },
  database: {
    cumulativeHitRate: 'database_buffer_cache_hit',
    hitsTotal: 'database_buffer_hits_total',
    readsTotal: 'database_buffer_reads_total'
  }
}

/** Metric names requested from the history endpoint to bootstrap a widget. */
export function bufferMetricNames(scope: BufferMetricScope): string[] {
  const metrics = BUFFER_METRICS[scope]

  return [metrics.cumulativeHitRate, metrics.hitsTotal, metrics.readsTotal]
}

/**
 * A pair of counter samples only becomes a rate when the samples are close
 * enough: after a longer gap (an inactive tab, a reconnect) the average would
 * describe a window nobody is looking at.
 */
export const MAX_RATE_WINDOW_MS = 15_000

/**
 * Gauge scale, two thresholds and three levels:
 * below DANGER is red, from HEALTHY is green, everything in between is amber.
 */
export const HIT_RATE_DANGER_PERCENT = 95
export const HIT_RATE_HEALTHY_PERCENT = 99

export type MetricSample = {
  value: number
  /** Unix milliseconds */
  timestamp: number
}

/** Builds the SSE endpoint URL, tolerating a trailing slash in the API base. */
export function metricsStreamUrl(baseUrl: string | undefined): string {
  const base = (baseUrl ?? '').replace(/\/+$/, '')

  return `${base}/metrics/stream`
}

/** Parses a single measurement of the SSE payload. */
export function parseMetricSample(payload: unknown): MetricSample | null {
  if (typeof payload !== 'object' || payload === null) return null

  const candidate = payload as { value?: unknown; timestamp?: unknown }

  const timestamp = typeof candidate.timestamp === 'string' ? Date.parse(candidate.timestamp) : NaN
  const value = typeof candidate.value === 'number' ? candidate.value : Number(candidate.value)

  if (!Number.isFinite(value) || !Number.isFinite(timestamp)) return null

  return { value, timestamp }
}

/** Parses an SSE `metrics` payload into measurements keyed by metric name. */
export function parseMetricsBatch(raw: string): Record<string, MetricSample> {
  let parsed: unknown

  try {
    parsed = JSON.parse(raw)
  } catch {
    return {}
  }

  if (typeof parsed !== 'object' || parsed === null) return {}

  const batch: Record<string, MetricSample> = {}

  for (const [name, payload] of Object.entries(parsed as Record<string, unknown>)) {
    const sample = parseMetricSample(payload)

    if (sample) batch[name] = sample
  }

  return batch
}

/**
 * Per second rate between two samples of a monotonic counter.
 *
 * Null means the pair cannot describe a rate: a non positive window, a window
 * longer than MAX_RATE_WINDOW_MS, or a counter that went backwards because
 * Postgres statistics were reset.
 */
export function counterRate(current: MetricSample, previous: MetricSample): number | null {
  const seconds = (current.timestamp - previous.timestamp) / 1000

  if (seconds <= 0 || seconds * 1000 > MAX_RATE_WINDOW_MS) return null
  if (current.value < previous.value) return 0

  return (current.value - previous.value) / seconds
}

/**
 * Hit rate of the last window: hits / (hits + misses) in percent.
 *
 * Null means "no traffic inside the window" and must not be rendered as 0%: on
 * an idle cluster there is nothing to be wrong about.
 */
export function intervalHitRate(
  hitsPerSecond: number | null,
  readsPerSecond: number | null
): number | null {
  if (hitsPerSecond === null || readsPerSecond === null) return null

  const total = hitsPerSecond + readsPerSecond

  if (total <= 0) return null

  return (hitsPerSecond / total) * 100
}

/** Half of the smallest compaction step, see formatRate. */
const RATE_PROMOTION = 999.5

/**
 * Formats a per second value the way the stats show it: whole numbers only,
 * because these are hit and miss counters, not fractions of a hit.
 *
 * Large values are compacted to an integer suffix: 12400 → '12k', 2500000 → '3M'.
 * The promotion threshold keeps rounding from printing '1000k' instead of '1M'.
 */
export function formatRate(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return '—'

  const rounded = Math.round(value)

  if (rounded <= 0) return '0'
  if (rounded >= RATE_PROMOTION * 1e6) return `${Math.round(rounded / 1e9)}B`
  if (rounded >= RATE_PROMOTION * 1e3) return `${Math.round(rounded / 1e6)}M`
  if (rounded >= RATE_PROMOTION) return `${Math.round(rounded / 1e3)}k`

  return String(rounded)
}

export type HitRateTone = 'bad' | 'warning' | 'good'

/**
 * Level of the gauge for a hit rate:
 *   percent < HIT_RATE_DANGER_PERCENT   → 'bad'     (red)
 *   percent < HIT_RATE_HEALTHY_PERCENT  → 'warning' (amber)
 *   otherwise                           → 'good'    (green)
 *
 * No interpolation on purpose: three states are easier to read and to explain
 * than a gradient, and the colours come from the design system.
 */
export function hitRateTone(percent: number): HitRateTone {
  if (percent < HIT_RATE_DANGER_PERCENT) return 'bad'
  if (percent < HIT_RATE_HEALTHY_PERCENT) return 'warning'

  return 'good'
}

export function clampPercent(percent: number): number {
  return clamp(percent, 0, 100)
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}
