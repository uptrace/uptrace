import { shallowRef, computed, watch, proxyRefs } from 'vue'

import { Values } from '@/use/router'

export interface Order {
  column: string
  desc: boolean
}

export type UseOrder = ReturnType<typeof useOrder>

export function useOrder() {
  const column = shallowRef<string>('')
  const desc = shallowRef(true)

  const axiosParams = shallowRef({})
  let watchPaused = false

  const icon = computed(() => {
    return desc.value ? 'mdi-arrow-down' : 'mdi-arrow-up'
  })

  watch(
    () => {
      const params: Record<string, any> = {}
      if (column.value) {
        params.sort_by = column.value
        params.sort_desc = desc.value
      }
      return params
    },
    (params) => {
      if (!watchPaused) {
        axiosParams.value = params
      }
    },
    { immediate: true, flush: 'sync' },
  )

  function queryParams() {
    if (column.value) {
      return {
        sort_by: column.value,
        sort_desc: desc.value,
      }
    }
    return {}
  }

  function parseQueryParams(queryParams: Values) {
    column.value = queryParams.string('sort_by')
    desc.value = queryParams.boolean('sort_desc', true)
  }

  function change(order: Order): boolean {
    if (column.value !== order.column || desc.value !== order.desc) {
      column.value = order.column
      desc.value = order.desc
      return true
    }
    return false
  }

  function toggle(columnValue: string): void {
    if (column.value === columnValue) {
      desc.value = !desc.value
      return
    }
    column.value = columnValue
    desc.value = true
  }

  function thClass(columnValue: string): string[] {
    const cls = ['cursor-pointer']
    if (column.value === columnValue) {
      cls.push('active')
    }
    if (desc.value) {
      cls.push('desc')
    } else {
      cls.push('asc')
    }
    return cls
  }

  function withPausedWatch(cb: () => void) {
    watchPaused = true
    const result = cb()
    watchPaused = false
    return result
  }

  function reset() {
    column.value = ''
    desc.value = true
  }

  return proxyRefs({
    column,
    desc,
    icon,

    axiosParams,
    withPausedWatch,

    change,
    reset,
    toggle,
    thClass,

    queryParams,
    parseQueryParams,
  })
}
