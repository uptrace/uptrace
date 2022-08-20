import { computed, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'
import { useForceReload } from '@/use/force-reload'

// Types
import { Metric, ActiveMetric, MetricAlias, Instrument } from '@/metrics/types'

export type UseMetrics = ReturnType<typeof useMetrics>

export function useMetrics() {
  const { route } = useRouter()
  const { forceReloadParams } = useForceReload()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}`,
      params: forceReloadParams.value,
    }
  })

  const metrics = computed((): Metric[] => {
    return data.value?.metrics ?? []
  })

  const noData = computed((): boolean => {
    return status.value.hasData() && metrics.value.length === 0
  })

  return proxyRefs({
    status,
    loading,
    noData,
    items: metrics,

    reload,
  })
}

export function metricShortName(name: string): string {
  const ident: string[] = []

  const ss = name.split(/[./]/).reverse()
  for (let s of ss) {
    s = s.replaceAll(/[^a-z0-9]+/gi, '_')
    ident.push(s)

    if (s.length >= 2 && s.match(/[a-z][a-z0-9_]+/i)) {
      break
    }
  }

  return ident.reverse().join('.')
}

//------------------------------------------------------------------------------

export function useActiveMetrics(
  metrics: ComputedRef<Metric[]>,
  metricAliases: ComputedRef<MetricAlias[]>,
) {
  return computed((): ActiveMetric[] => {
    const active: ActiveMetric[] = []

    for (let v of metricAliases.value) {
      const found = metrics.value.find((m) => m.name === v.name)
      if (found) {
        active.push({
          ...found,
          alias: v.alias,
        })
      }
    }

    return active
  })
}

export function defaultMetricColumn(instrument: Instrument, alias: string) {
  alias = '$' + alias
  switch (instrument) {
    case Instrument.Gauge:
    case Instrument.Additive:
      return alias
    case Instrument.Counter:
      return `per_min(${alias})`
    case Instrument.Histogram:
      return `p50(${alias})`
    default:
      // eslint-disable-next-line no-console
      console.error('unknown metric instrument', instrument)
      return alias
  }
}
