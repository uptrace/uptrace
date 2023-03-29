import { omit } from 'lodash-es'
import { shallowRef, computed, proxyRefs } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'

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

  const q = shallowRef('')
  const selected = shallowRef<Record<string, string[]>>({})

  const selectedLength = computed((): number => {
    return Object.keys(selected.value).length
  })

  const axiosParams = computed((): Record<string, any> => {
    const params: Record<string, any> = {}
    if (q.value) {
      params.q = q.value
    }
    for (let key in selected.value) {
      params[`attrs[${key}]`] = selected.value[key]
    }
    return params
  })

  function select(item: FacetItem) {
    selected.value = {
      ...selected.value,
      [item.key]: [item.value],
    }
  }

  function reset(item: FacetItem) {
    const value = selected.value[item.key]

    if (value && value.includes(item.value)) {
      if (value.length === 1) {
        selected.value = omit(selected.value, item.key)
        return
      }
    }

    selected.value = {
      ...selected.value,
      [item.key]: [item.value],
    }
  }

  function toggle(item: FacetItem) {
    let value = selected.value[item.key]

    if (value && value.includes(item.value)) {
      selected.value = omit(selected.value, item.key)
      return
    }

    if (value) {
      value.push(item.value)
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

  useRouteQuery().sync({
    fromQuery(query) {
      if ('q' in query) {
        q.value = query.q
      }

      selected.value = {}

      for (let key in query) {
        if (!key.startsWith(queryPrefix)) {
          continue
        }
        const value = toArray(query[key])
        if (value && value.length) {
          selected.value[key.slice(queryPrefix.length)] = value
        }
      }
    },

    toQuery() {
      const query: Record<string, any> = {
        q: q.value,
      }
      for (let key in selected.value) {
        query[queryPrefix + key] = selected.value[key]
      }
      return query
    },
  })

  return proxyRefs({
    q,
    selected,
    selectedLength,
    axiosParams,

    isSelected,
    select,
    reset,
    toggle,
    resetAll,
  })
}

function toArray(v: any): string[] {
  if (Array.isArray(v)) {
    return v
  }
  if (typeof v === 'string') {
    return [v]
  }
  return []
}
