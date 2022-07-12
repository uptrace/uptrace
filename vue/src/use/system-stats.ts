import { orderBy } from 'lodash'
import { ref, computed, watch, proxyRefs } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { usePager } from '@/use/pager'
import { useOrder } from '@/use/order'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

// Utilities
import { splitTypeSystem } from '@/models/otelattr'

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

interface TypeItem {
  type: string
  numSystem: number
}

export type UseSystemStats = ReturnType<typeof useSystemStats>

export function useSystemStats(dateRange: UseDateRange) {
  const { route } = useRouter()
  const pager = usePager({ perPage: 15 })
  const order = useOrder({ column: 'system', desc: false })
  const filters = ref([])

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/tracing/${projectId}/systems-stats`,
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
    if (!filters.value.length) {
      return sortedSystems.value
    }

    return sortedSystems.value.filter((sys) => {
      for (let filter of filters.value) {
        if (sys.system.startsWith(filter)) {
          return true
        }
      }

      return false
    })
  })

  const pageSystems = computed((): OverviewItem[] => {
    const pageSystems = filteredSystems.value.slice(pager.pos.start, pager.pos.end)
    return pageSystems
  })

  const types = computed((): TypeItem[] => {
    const typeMap: Record<string, TypeItem> = {}

    for (let sys of systems.value) {
      const [type] = splitTypeSystem(sys.system)
      let typeItem = typeMap[type]
      if (!typeItem) {
        typeItem = {
          type,
          numSystem: 0,
        }
        typeMap[type] = typeItem
      }

      typeItem.numSystem++
    }

    const types: TypeItem[] = []

    for (let type in typeMap) {
      types.push(typeMap[type])
    }

    orderBy(types, 'type')
    return types
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
    filters,

    loading,
    list: filteredSystems,
    pageSystems,
    types,
  })
}
