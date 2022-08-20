import { shallowRef, computed, proxyRefs, watch } from 'vue'

import { useRouteQuery } from '@/use/router'

export interface Order {
  column: string
  desc: boolean
}

export interface OrderConfig extends Partial<Order> {
  syncQuery?: boolean
}

export type UseOrder = ReturnType<typeof useOrder>

export function useOrder(cfg: OrderConfig = {}) {
  cfg.column = cfg.column ?? ''
  cfg.desc = cfg.desc ?? true
  cfg.syncQuery = cfg.syncQuery ?? false

  const column = shallowRef<string | undefined>(cfg.column)
  const desc = shallowRef(cfg.desc)

  const axiosParamsLocked = shallowRef(false)
  const axiosParams = shallowRef<Record<string, any>>({})

  const icon = computed(() => {
    return desc.value ? 'mdi-arrow-down' : 'mdi-arrow-up'
  })

  watch(
    () => {
      return {
        sort_by: column.value,
        sort_desc: desc.value,
      }
    },
    (params) => {
      if (!axiosParamsLocked.value) {
        axiosParams.value = params
      }
    },
    { immediate: true, flush: 'sync' },
  )

  if (cfg.syncQuery) {
    useRouteQuery().sync({
      fromQuery(params) {
        if (params.sort_by) {
          column.value = params.sort_by
          desc.value = params.sort_desc === '1'
        }
      },
      toQuery() {
        if (column.value) {
          return {
            sort_by: column.value,
            sort_desc: desc.value ? '1' : '0',
          }
        }
        return {}
      },
    })
  }

  function change(order: Order): void {
    column.value = order.column
    desc.value = order.desc
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

  function lockAxiosParams() {
    axiosParamsLocked.value = true
  }

  function unlockAxiosParams() {
    axiosParamsLocked.value = false
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
    lockAxiosParams,
    unlockAxiosParams,
    withLockedAxiosParams,

    change,
    reset,
    toggle,
    thClass,
  })
}
