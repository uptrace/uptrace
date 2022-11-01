import { shallowRef, computed, watch, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'

export const CATEGORY_CORE = 'core'

export interface Item {
  text: string
  count: number
  visible: boolean
}

export function useSpanAttrs(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()

  const { status, loading, data } = useWatchAxios(() => {
    const params = axiosParamsSource()
    if (!params) {
      return params
    }

    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/suggestions/attributes?kind=text`,
      params,
    }
  })

  const items = computed((): Item[] => {
    return data.value?.suggestions ?? []
  })

  return proxyRefs({
    status,
    loading,
    items,
  })
}

export function useSpanAttrValues(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()
  const searchQuery = shallowRef('')
  const hasMore = shallowRef(true)

  const { status, loading, data } = useWatchAxios(() => {
    let params = axiosParamsSource()
    if (!params) {
      return params
    }

    const { projectId } = route.value.params
    params = {
      ...params,
      attr_value: searchQuery.value,
    }
    if (!hasMore.value) {
      params.$ignore_attr_value = true
    }

    return {
      url: `/api/v1/tracing/${projectId}/suggestions/values`,
      params,
      debounce: 500,
    }
  })

  const items = computed((): Item[] => {
    return data.value?.suggestions ?? []
  })

  const filteredItems = computed(() => {
    if (hasMore.value) {
      return items.value
    }
    return items.value.filter((item) => {
      return item.text.includes(searchQuery.value ?? '')
    })
  })

  watch(
    () => data.value?.hasMore,
    (hasMoreValue) => {
      hasMore.value = hasMoreValue
    },
  )

  return proxyRefs({
    searchQuery,

    status,
    loading,
    items: filteredItems,
    hasMore,
  })
}
