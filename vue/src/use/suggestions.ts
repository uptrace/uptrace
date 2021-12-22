import { shallowRef, computed, proxyRefs } from '@vue/composition-api'

// Composables
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export interface Suggestion {
  text: string
  hint?: string
}

export type UseSuggestions = ReturnType<typeof useSuggestions>

interface Config {
  suggestSearchInput?: boolean
}

export function useSuggestions<T extends Suggestion>(
  reqSource: AxiosRequestSource,
  cfg: Config = {},
) {
  const searchInput = shallowRef('')
  const suggestSearchInput = cfg.suggestSearchInput ?? false

  const { loading, data } = useWatchAxios(() => {
    return reqSource()
  })

  const items = computed((): T[] => {
    return data.value?.suggestions ?? []
  })

  const filteredItems = computed((): T[] => {
    let filtered = items.value.slice()

    if (!searchInput.value) {
      return filtered
    }

    filtered = filtered.filter((item) => item.text.indexOf(searchInput.value) >= 0)

    if (suggestSearchInput) {
      const item = { text: searchInput.value }
      if (filtered.length <= 5) {
        filtered.push(item as T)
      } else {
        filtered.unshift(item as T)
      }
    }

    return filtered
  })

  return proxyRefs({
    searchInput,

    loading,
    data,
    items,
    filteredItems,
  })
}
