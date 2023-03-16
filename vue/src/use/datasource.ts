import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { shallowRef, computed, watch, proxyRefs } from 'vue'

import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export interface Item {
  value: string
  text: string
}

interface Config {
  dataKey?: string
  suggestSearchInput?: boolean
}

export type UseDataSource = ReturnType<typeof useDataSource>

export function useDataSource<T extends Item>(
  axiosReqSource: AxiosRequestSource,
  conf: Config = {},
) {
  return proxyRefs(useDataSourceRefs<T>(axiosReqSource, conf))
}

export function useDataSourceRefs<T extends Item>(
  axiosReqSource: AxiosRequestSource,
  conf: Config = {},
) {
  const dataKey = conf.dataKey ?? 'items'
  const suggestSearchInput = conf.suggestSearchInput ?? false

  const searchInput = shallowRef('')
  const hasMore = shallowRef(false)

  const { status, loading, data, errorMessage, reload } = useWatchAxios(() => {
    const req = axiosReqSource()
    if (!req) {
      return req
    }

    req.params ??= {}
    req.params.search_input = searchInput.value
    if (!hasMore.value) {
      req.params.$ignore_search_input = true
    }

    return req
  })

  const items = computed((): T[] => {
    const items: T[] = data.value?.[dataKey] ?? []
    return items.map((item) => {
      if (!item.text) {
        item.text = item.value
      }
      return item
    })
  })

  const filteredItems = computed((): T[] => {
    let filtered = items.value.slice()

    if (!searchInput.value) {
      return filtered
    }

    if (!hasMore.value) {
      // @ts-ignore
      filtered = fuzzyFilter(filtered, searchInput.value, { key: 'text' })
    }

    if (suggestSearchInput) {
      const item = { value: searchInput.value }
      if (filtered.length <= 5) {
        filtered.push(item as T)
      } else {
        filtered.unshift(item as T)
      }
    }

    return filtered
  })

  const values = computed((): string[] => {
    return items.value.map((item) => item.value)
  })

  const filteredValues = computed((): string[] => {
    return filteredItems.value.map((item) => item.value)
  })

  watch(
    () => data.value?.hasMore ?? false,
    (hasMoreValue) => {
      hasMore.value = hasMoreValue
    },
  )

  return {
    searchInput,
    errorMessages: errorMessage,

    status,
    loading,
    data,
    reload,

    items,
    filteredItems,

    values,
    filteredValues,
  }
}
