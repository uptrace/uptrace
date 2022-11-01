import { shallowRef, computed, proxyRefs } from 'vue'

// Composables
import { useRoute, useRouteQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

export interface Item {
  value: string
  count: number
}

export type UseEnvs = ReturnType<typeof useEnvs>

export function useEnvs(dateRange: UseDateRange) {
  const stickyFilter = useQuickSpanFilter('envs', dateRange)

  return proxyRefs({
    ...stickyFilter,
  })
}

export type UseServices = ReturnType<typeof useServices>

export function useServices(dateRange: UseDateRange) {
  const stickyFilter = useQuickSpanFilter('services', dateRange)

  return proxyRefs({
    ...stickyFilter,
  })
}

function useQuickSpanFilter(id: string, dateRange: UseDateRange) {
  const route = useRoute()
  const lastActive = shallowRef<string[]>([])

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/${id}`,
      params: {
        ...dateRange.axiosParams(),
      },
      cache: true,
    }
  })

  const items = computed((): Item[] => {
    return data.value?.items ?? []
  })

  const active = computed({
    get() {
      return lastActive.value.filter(
        (value) => items.value.findIndex((item) => item.value === value) >= 0,
      )
    },
    set(items) {
      lastActive.value = items
    },
  })

  useRouteQuery().sync({
    fromQuery(params) {
      if (!Object.keys(params).length) {
        return
      }

      const paramValue = params[id]

      if (!paramValue) {
        active.value = []
        return
      }

      if (Array.isArray(paramValue)) {
        active.value = paramValue
      } else if (typeof paramValue === 'string') {
        active.value = [paramValue]
      }
    },
    toQuery() {
      if (lastActive.value.length) {
        return { [id]: lastActive.value }
      }
      return {}
    },
  })

  function axiosParams() {
    if (!status.value.hasData()) {
      return {
        [id]: undefined,
      }
    }
    if (active.value.length) {
      return {
        [id]: active.value,
      }
    }
    return {}
  }

  return {
    status,
    loading,
    items,
    active,

    axiosParams,
    reload,
  }
}
