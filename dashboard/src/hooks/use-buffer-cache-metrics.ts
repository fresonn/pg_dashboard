import { useEffect, useRef, useState } from 'react'
import { metricsHistory } from '@/lib/api/gen'
import {
  BUFFER_METRIC,
  BUFFER_METRIC_NAMES,
  counterRate,
  intervalHitRate,
  metricsStreamUrl,
  parseMetricSample,
  parseMetricsBatch,
  type MetricSample
} from '@/lib/metrics'

export type MetricsStatus = 'connecting' | 'live' | 'offline'

export type BufferCacheMetrics = {
  /** Cache hit rate of the last collection window, in percent */
  hitRate: number | null
  /**
   * Last hit rate that could be calculated.
   *
   * A window without traffic (no hits and no misses) has no hit rate, which must
   * not empty the gauge: an idle cluster says nothing about the cache, so the
   * last measured value stays on screen until a new one arrives.
   */
  lastKnownHitRate: number | null
  /** Cumulative hit ratio reported by Postgres since the last statistics reset */
  cumulativeHitRate: number | null
  hitsPerSecond: number | null
  readsPerSecond: number | null
  /** Actual width of the last rate window (milliseconds) */
  windowMs: number | null
  /** Timestamp of the last measurement (unix milliseconds) */
  updatedAt: number | null
  status: MetricsStatus
}

const INITIAL_STATE: BufferCacheMetrics = {
  hitRate: null,
  lastKnownHitRate: null,
  cumulativeHitRate: null,
  hitsPerSecond: null,
  readsPerSecond: null,
  windowMs: null,
  updatedAt: null,
  status: 'connecting'
}

/** How many stored points are enough to bootstrap the counters (current + previous). */
const HISTORY_LIMIT = 2

type Counter = 'hits' | 'reads'

type Model = {
  status: MetricsStatus
  counters: Record<Counter, MetricSample | null>
  rates: Record<Counter, number | null>
  cumulative: MetricSample | null
  lastHitRate: number | null
  windowMs: number | null
  updatedAt: number | null
}

function createModel(): Model {
  return {
    status: 'connecting',
    counters: { hits: null, reads: null },
    rates: { hits: null, reads: null },
    cumulative: null,
    lastHitRate: null,
    windowMs: null,
    updatedAt: null
  }
}

/**
 * Live shared buffers metrics for the Cluster Overview widget.
 *
 * Data flow:
 *  1. the history endpoint seeds the counters, so the first rendered rate is a
 *     real measurement instead of waiting for a second live sample;
 *  2. one SSE connection delivers every metric; samples with an older timestamp
 *     than the already applied one are ignored, which makes the history seed and
 *     the live stream idempotent (no duplicated or replayed points);
 *  3. after a reconnect the counters are re-seeded from history, which closes
 *     the gap the browser could not see while the connection was down.
 *
 * The rates are derived from the cumulative counters published by the backend:
 * hits/s and reads/s are deltas over the actual collection interval.
 */
export function useBufferCacheMetrics(): BufferCacheMetrics {
  const [state, setState] = useState<BufferCacheMetrics>(INITIAL_STATE)

  const model = useRef<Model>(createModel())

  useEffect(() => {
    let disposed = false

    const abort = new AbortController()

    const flush = () => {
      if (disposed) return

      const current = model.current
      const hitRate = intervalHitRate(current.rates.hits, current.rates.reads)

      // Remeasure, but never forget: an empty window keeps the previous value.
      if (hitRate !== null) current.lastHitRate = hitRate

      setState({
        status: current.status,
        hitRate,
        lastKnownHitRate: current.lastHitRate,
        cumulativeHitRate: current.cumulative?.value ?? null,
        hitsPerSecond: current.rates.hits,
        readsPerSecond: current.rates.reads,
        windowMs: current.windowMs,
        updatedAt: current.updatedAt
      })
    }

    const applySample = (name: string, sample: MetricSample) => {
      const current = model.current

      if (name === BUFFER_METRIC.hitsTotal || name === BUFFER_METRIC.readsTotal) {
        const counter: Counter = name === BUFFER_METRIC.hitsTotal ? 'hits' : 'reads'
        const previous = current.counters[counter]

        // Monotonic per metric: history seeding and the live snapshot may
        // overlap, the older sample must not win.
        if (previous && previous.timestamp > sample.timestamp) return

        current.rates[counter] = previous ? counterRate(sample, previous) : null
        current.counters[counter] = sample

        if (previous) current.windowMs = sample.timestamp - previous.timestamp
      } else if (name === BUFFER_METRIC.cumulativeHitRate) {
        if (current.cumulative && current.cumulative.timestamp > sample.timestamp) return

        current.cumulative = sample
      } else {
        return
      }

      current.updatedAt = Math.max(current.updatedAt ?? 0, sample.timestamp)
    }

    const samplesOf = (
      series: Record<string, { timestamp: string; value: number }[]>,
      name: string
    ) =>
      (series[name] ?? [])
        .map((point) => parseMetricSample(point))
        .filter((sample): sample is MetricSample => sample !== null)
        .sort((left, right) => left.timestamp - right.timestamp)

    const seedCounter = (counter: Counter, samples: MetricSample[]) => {
      const current = model.current

      if (samples.length === 0) return

      const latest = samples[samples.length - 1]
      const previous = samples[samples.length - 2]

      if (current.counters[counter] && current.counters[counter]!.timestamp >= latest.timestamp)
        return

      current.rates[counter] = previous ? counterRate(latest, previous) : null
      current.counters[counter] = latest

      if (previous) current.windowMs = latest.timestamp - previous.timestamp
    }

    // Best effort: the live stream works without history, it just needs one
    // extra sample before the tiles show a rate.
    const seedFromHistory = async () => {
      try {
        const history = await metricsHistory(
          { names: BUFFER_METRIC_NAMES.join(','), limit: HISTORY_LIMIT },
          { signal: abort.signal }
        )

        if (disposed) return

        const series = history?.series ?? {}

        seedCounter('hits', samplesOf(series, BUFFER_METRIC.hitsTotal))
        seedCounter('reads', samplesOf(series, BUFFER_METRIC.readsTotal))

        const cumulative = samplesOf(series, BUFFER_METRIC.cumulativeHitRate)

        if (cumulative.length > 0) {
          applySample(BUFFER_METRIC.cumulativeHitRate, cumulative[cumulative.length - 1])
        }

        flush()
      } catch (error) {
        if (disposed || abort.signal.aborted) return

        console.warn('[metrics] history bootstrap failed, waiting for live samples', error)
      }
    }

    const source = new EventSource(metricsStreamUrl(import.meta.env.VITE_API_URL))

    source.addEventListener('open', () => {
      const reconnected = model.current.status === 'offline'

      model.current.status = 'live'
      flush()

      // A reconnect may have skipped points: history is the source of truth for
      // everything the stream could not deliver.
      if (reconnected) void seedFromHistory()
    })

    source.addEventListener('metrics', (event) => {
      const batch = parseMetricsBatch((event as MessageEvent<string>).data)

      for (const [name, sample] of Object.entries(batch)) {
        applySample(name, sample)
      }

      model.current.status = 'live'
      flush()
    })

    source.addEventListener('error', () => {
      // EventSource reconnects on its own; the hook only reflects the state and
      // refills the gap once the connection is back.
      model.current.status = 'offline'
      flush()
    })

    void seedFromHistory()

    return () => {
      disposed = true
      abort.abort()
      source.close()
    }
  }, [])

  return state
}
