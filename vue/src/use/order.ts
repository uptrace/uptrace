import { shallowRef, computed, watch, proxyRefs } from 'vue'

import { useRouteQuery } from '@/use/router'

export interface Order {
  column: string
  desc: boolean
}

export type UseOrder = ReturnType<typeof useOrder>

export function useOrder(conf: Partial<Order> = {}) {
  conf.column = conf.column ?? ''
  conf.desc = conf.desc ?? true

  const column = shallowRef<string | undefined>(conf.column)
  const desc = shallowRef(conf.desc)

  const axiosParams = shallowRef({})
  const axiosParamsLocked = shallowRef(false)
  const ignoreAxiosParamsEnabled = shallowRef(false)

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
      if (ignoreAxiosParamsEnabled.value) {
        params.$ignore_sort_by = true
        params.$ignore_sort_desc = true
      }
      return params
    },
    (params) => {
      if (!axiosParamsLocked.value) {
        axiosParams.value = params
      }
    },
    { immediate: true, flush: 'sync' },
  )

  function syncQueryParams() {
    useRouteQuery().sync({
      fromQuery(params) {
        if (params.sort_by) {
          column.value = params.sort_by
          desc.value = params.sort_desc === '1'
        }
      },
      toQuery() {
        return queryParams()
      },
    })
  }

  function queryParams() {
    if (column.value) {
      return {
        sort_by: column.value,
        sort_desc: desc.value ? '1' : '0',
      }
    }
    return {}
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

  function withLockedAxiosParams(cb: () => void) {
    const oldValue = axiosParamsLocked.value
    axiosParamsLocked.value = true

    cb()

    axiosParamsLocked.value = oldValue
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
    ignoreAxiosParamsEnabled,
    withLockedAxiosParams,

    syncQueryParams,
    queryParams,
    change,
    reset,
    toggle,
    thClass,
  })
}
