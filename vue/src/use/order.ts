import { shallowRef, computed, proxyRefs, watch } from '@vue/composition-api'

export interface Order {
  column: string
  desc: boolean
}

export type UseOrder = ReturnType<typeof useOrder>

export function useOrder(cfg: Partial<Order> = {}) {
  cfg.column = cfg.column ?? ''
  cfg.desc = cfg.desc ?? true

  const column = shallowRef<string | undefined>(cfg.column)
  const desc = shallowRef(cfg.desc)
  const axiosParams = shallowRef<Record<string, any>>({})

  const icon = computed(() => {
    return desc.value ? 'mdi-arrow-down' : 'mdi-arrow-up'
  })

  watch(
    () => {
      return {
        sort_by: column.value,
        sort_dir: descAsc(desc.value),
      }
    },
    (params) => {
      axiosParams.value = params
    },
    { immediate: true, flush: 'sync' },
  )

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

  function reset() {
    column.value = ''
    desc.value = true
  }

  return proxyRefs({
    column,
    desc,
    icon,

    axiosParams,

    change,
    reset,
    toggle,
    thClass,
  })
}

function descAsc(isDesc: boolean): string {
  return isDesc ? 'desc' : 'asc'
}
