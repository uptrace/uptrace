import { orderBy } from 'lodash-es'
import { computed, watch, proxyRefs } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'
import { usePager } from '@/use/pager'
import { useOrder } from '@/use/order'

interface Group {
  attr: any

  count: number
  rate: number

  errorCount?: number
  errorPct?: number
  p50Duration?: number
  p90Duration?: number
  p99Duration?: number
  maxDuration?: number

  stats: {
    count: number[]
    rate: number[]

    errorCount?: number[]
    errorPct?: number[]
    p50Duration?: number[]
    p90Duration?: number[]
    p99Duration?: number[]
    maxDuration?: number[]
  }
}

export function useOverview(params: () => Record<string, any>) {
  const { route } = useRouter()
  const pager = usePager({ perPage: 15 })
  const order = useOrder({ column: 'attr', desc: false })

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/overview`,
      params: params(),
    }
  })

  const groups = computed((): Group[] => {
    const groups = data.value?.groups ?? []
    return groups
  })

  const sortedGroups = computed((): Group[] => {
    const sortedGroups = orderBy(groups.value, order.column ?? 'attr', order.desc ? 'desc' : 'asc')
    return sortedGroups
  })

  const pagedGroups = computed((): Group[] => {
    const pagedGroups = sortedGroups.value.slice(pager.pos.start, pager.pos.end)
    return pagedGroups
  })

  watch(
    groups,
    (groups) => {
      pager.numItem = groups.length
    },
    { flush: 'sync' },
  )

  return proxyRefs({
    pager,
    order,

    loading,
    groups: sortedGroups,
    pagedGroups,
  })
}
