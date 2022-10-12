import { computed, proxyRefs } from 'vue'

// Composables
import { useStorage } from '@/use/local-storage'
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

export type UseEnvs = ReturnType<typeof useEnvs>

export function useEnvs(dateRange: UseDateRange) {
  const stickyFilter = useStickyFilter('envs', dateRange)

  return proxyRefs({
    ...stickyFilter,
  })
}

export type UseServices = ReturnType<typeof useServices>

export function useServices(dateRange: UseDateRange) {
  const stickyFilter = useStickyFilter('services', dateRange)

  return proxyRefs({
    ...stickyFilter,
  })
}

function useStickyFilter(id: string, dateRange: UseDateRange) {
  const route = useRoute()

  const { item: lastActive } = useStorage<string[]>(
    computed(() => {
      return `${id}:active:${route.value.params.projectId}`
    }),
    [],
  )

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

  const items = computed(() => {
    return data.value?.items ?? []
  })

  const active = computed({
    get() {
      return lastActive.value.filter((item) => items.value.indexOf(item) >= 0)
    },
    set(items) {
      lastActive.value = items
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
