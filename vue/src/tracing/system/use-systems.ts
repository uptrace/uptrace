import { cloneDeep, orderBy } from 'lodash-es'
import { shallowRef, computed, proxyRefs } from 'vue'

// Composables
import { useRoute, useRouteQuery } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

export interface System {
  projectId: number | string
  system: string

  count: number
  rate: number
  errorCount: number
  errorPct: number

  indent?: boolean
}

export type SystemsFilter = (systems: System[]) => System[]

export interface SystemTreeNode extends System {
  numChildren: number
  children?: SystemTreeNode[]
}

export type UseSystems = ReturnType<typeof useSystems>

export function useSystems(params: () => Record<string, any>) {
  const route = useRoute()

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/systems`,
      params: params(),
    }
  })

  const systems = computed((): System[] => {
    const systems = data.value?.systems ?? []
    return addDummySystems(systems ?? [])
  })

  const hasNoData = computed(() => {
    return data.value?.hasNoData ?? false
  })

  const internalValue = shallowRef<string[]>([])

  const activeSystem = computed({
    get(): string[] {
      return internalValue.value
    },
    set(system: string | string[]) {
      if (Array.isArray(system)) {
        internalValue.value = system
      } else if (system) {
        internalValue.value = [system]
      } else {
        internalValue.value = []
      }
    },
  })

  function syncQueryParams() {
    useRouteQuery().sync({
      fromQuery(params) {
        const system = params.system
        if (system) {
          activeSystem.value = system
        } else {
          // Reset so we can pick a new system from the list later.
          activeSystem.value = []
        }
      },
      toQuery() {
        if (activeSystem.value.length) {
          return { system: activeSystem.value }
        }
      },
    })
  }

  function axiosParams() {
    return {
      system: activeSystem.value,
    }
  }

  function queryParams() {
    return {
      system: activeSystem.value,
    }
  }

  function reset(): void {
    activeSystem.value = []
  }

  return proxyRefs({
    loading,
    items: systems,
    hasNoData,
    activeSystem,

    syncQueryParams,
    axiosParams,
    queryParams,
    reset,
  })
}

function addDummySystems(systems: System[]): System[] {
  if (!systems.length) {
    return []
  }
  systems = cloneDeep(systems)

  const typeMap: Record<string, SystemTreeNode> = {}

  for (let sys of systems) {
    let typ = sys.system

    const i = typ.indexOf(':')
    if (i >= 0) {
      typ = typ.slice(0, i)
    }

    const typeSys = typeMap[typ]
    if (typeSys) {
      typeSys.count += sys.count
      typeSys.rate += sys.rate
      typeSys.numChildren!++
      continue
    }

    typeMap[typ] = {
      ...sys,
      system: typ,
      numChildren: 1,
    }
  }

  for (let systemType in typeMap) {
    const typ = typeMap[systemType]
    if (typ.numChildren <= 1) {
      continue
    }

    for (let item of systems) {
      if (item.system.startsWith(typ.system)) {
        item.indent = true
      }
    }

    systems.push({
      ...typ,
      system: typ.system + ':all',
    })
  }

  systems = orderBy(systems, 'system')
  return systems
}
