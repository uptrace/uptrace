import { shallowRef, computed, proxyRefs } from 'vue'
import { refDebounced } from '@/use/ref-debounced'

// Composables
import { Values } from '@/use/router'

export interface Facet {
  key: string
  items: FacetItem[]
}

export interface FacetItem {
  key: string
  value: string
  count: number
}

export type UseFacetedSearch = ReturnType<typeof useFacetedSearch>

export function useFacetedSearch() {
  const queryPrefix = 'attrs.'

  const searchInput = shallowRef('')
  const debouncedSearchInput = refDebounced(searchInput, 600)
  const selected = shallowRef<Record<string, string[]>>({})

  const selectedLength = computed((): number => {
    return Object.keys(selected.value).length
  })

  function select(item: FacetItem) {
    selected.value = {
      ...selected.value,
      [item.key]: [item.value],
    }
  }

  function toggle(item: FacetItem) {
    let value = selected.value[item.key]

    if (value) {
      const idx = value.indexOf(item.value)
      if (idx >= 0) {
        value.splice(idx, 1)
      } else {
        value.push(item.value)
      }
    } else {
      value = [item.value]
    }

    selected.value = {
      ...selected.value,
      [item.key]: value,
    }
  }

  function toggleOne(item: FacetItem) {
    let value = selected.value[item.key]

    if (value && value.length === 1 && value.includes(item.value)) {
      value = []
    } else {
      value = [item.value]
    }

    selected.value = {
      ...selected.value,
      [item.key]: value,
    }
  }

  function isSelected(item: FacetItem): boolean {
    const value = selected.value[item.key]
    return value && value.includes(item.value)
  }

  function resetAll() {
    selected.value = {}
  }

  function axiosParams() {
    const params: Record<string, any> = {}
    if (debouncedSearchInput.value) {
      params.q = debouncedSearchInput.value
    }
    for (let key in selected.value) {
      params[`attrs[${key}]`] = selected.value[key]
    }
    return params
  }

  function queryParams() {
    const queryParams: Record<string, any> = {
      q: debouncedSearchInput.value,
    }
    for (let key in selected.value) {
      queryParams[queryPrefix + key] = selected.value[key]
    }
    return queryParams
  }

  function parseQueryParams(queryParams: Values) {
    searchInput.value = queryParams.string('q')
    debouncedSearchInput.flush()

    selected.value = {}
    queryParams.forEach((key, value) => {
      if (!key.startsWith(queryPrefix)) {
        return
      }
      key = key.slice(queryPrefix.length)
      selected.value[key] = value
    })
  }

  return proxyRefs({
    searchInput,
    selected,
    selectedLength,

    isSelected,
    select,
    toggle,
    toggleOne,
    resetAll,

    axiosParams,
    queryParams,
    parseQueryParams,
  })
}
