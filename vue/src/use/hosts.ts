import { orderBy } from 'lodash'
import { computed, watch, proxyRefs } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'
import { UseDateRange } from '@/use/date-range'
import { usePager } from '@/use/pager'
import { useOrder } from '@/use/order'
import { UseSystems } from '@/use/systems'

// Utilities
import { xkey } from '@/models/otelattr'

export interface OverviewItem {
  system: string

  count: number
  rate: number
  errorCount: number
  errorPct: number

  p50: number
  p90: number
  p99: number

  stats: Stats
}

interface Stats {
  count: number[]
  errorCount: number[]
  p50: number[]
  p90: number[]
  p99: number[]
  time: string[]
}

export function useHosts(dateRange: UseDateRange, systems: UseSystems) {
  const { route } = useRouter()
  const pager = usePager({ perPage: 15 })
  const order = useOrder({ column: xkey.hostName, desc: false })

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/hosts`,
      params: {
        ...systems.axiosParams(),
        ...dateRange.axiosParams(),
      },
    }
  })

  const hosts = computed((): OverviewItem[] => {
    return data.value?.hosts ?? []
  })

  const sortedHosts = computed((): OverviewItem[] => {
    const sortedHosts = orderBy(
      hosts.value,
      order.column ?? xkey.hostName,
      order.desc ? 'desc' : 'asc',
    )
    return sortedHosts
  })

  const pageHosts = computed((): OverviewItem[] => {
    const pageHosts = sortedHosts.value.slice(pager.pos.start, pager.pos.end)
    return pageHosts
  })

  watch(
    hosts,
    (hosts) => {
      pager.numItem = hosts.length
    },
    { flush: 'sync' },
  )

  return proxyRefs({
    pager,
    order,

    loading,
    list: sortedHosts,
    pageHosts,
  })
}
