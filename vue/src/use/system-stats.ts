import { orderBy } from 'lodash'
import { ref, computed, watch, proxyRefs } from '@vue/composition-api'

// Composables
import { usePager } from '@/use/pager'
import { useOrder } from '@/use/order'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

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

export type UseSystemStats = ReturnType<typeof useSystemStats>

export function useSystemStats(dateRange: UseDateRange) {
  const pager = usePager({ perPage: 15 })
  const order = useOrder({ column: 'system', desc: false })
  const filter = ref('')

  const { loading, data } = useWatchAxios(() => {
    return {
      url: `/api/tracing/systems-stats`,
      params: {
        ...dateRange.axiosParams(),
      },
    }
  })

  const systems = computed((): OverviewItem[] => {
    const systems: OverviewItem[] = data.value?.systems ?? []
    return systems
  })

  const sortedSystems = computed((): OverviewItem[] => {
    const sortedSystems = orderBy(
      systems.value,
      order.column ?? 'system',
      order.desc ? 'desc' : 'asc',
    )
    return sortedSystems
  })

  const filteredSystems = computed((): OverviewItem[] => {
    if (!filter.value) {
      return sortedSystems.value
    }

    return sortedSystems.value.filter((sys) => {
      return sys.system.startsWith(filter.value)
    })
  })

  const pageSystems = computed((): OverviewItem[] => {
    const pageSystems = filteredSystems.value.slice(pager.pos.start, pager.pos.end)
    return pageSystems
  })

  const types = computed(() => {
    const types = []

    for (let sys of systems.value) {
      let typ = sys.system

      const i = typ.indexOf(':')
      if (i >= 0) {
        typ = typ.slice(0, i)
      }

      if (types.indexOf(typ) === -1) {
        types.push(typ)
      }
    }

    return types.sort()
  })

  watch(
    filteredSystems,
    (systems) => {
      pager.numItem = systems.length
    },
    { flush: 'sync' },
  )

  return proxyRefs({
    pager,
    order,
    filter,

    loading,
    list: filteredSystems,
    pageSystems,
    types,
  })
}
