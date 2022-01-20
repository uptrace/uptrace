import { orderBy } from 'lodash'
import { computed, watch, proxyRefs } from '@vue/composition-api'

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

export function useServices(dateRange: UseDateRange, systems: UseSystems) {
  const { route } = useRouter()
  const pager = usePager({ perPage: 15 })
  const order = useOrder({ column: xkey.serviceName, desc: false })

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/tracing/${projectId}/services`,
      params: {
        ...systems.axiosParams(),
        ...dateRange.axiosParams(),
      },
    }
  })

  const services = computed((): OverviewItem[] => {
    return data.value?.services ?? []
  })

  const sortedServices = computed((): OverviewItem[] => {
    const sortedServices = orderBy(
      services.value,
      order.column ?? xkey.serviceName,
      order.desc ? 'desc' : 'asc',
    )
    return sortedServices
  })

  const pageServices = computed((): OverviewItem[] => {
    const pageServices = sortedServices.value.slice(pager.pos.start, pager.pos.end)
    return pageServices
  })

  watch(
    services,
    (services) => {
      pager.numItem = services.length
    },
    { flush: 'sync' },
  )

  return proxyRefs({
    pager,
    order,

    loading,
    list: sortedServices,
    pageServices,
  })
}
