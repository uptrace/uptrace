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
}

export type SystemsFilter = (systems: System[]) => System[]

export interface SystemTreeNode extends System {
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

  function syncQuery() {
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

    syncQuery,
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
    const i = sys.system.indexOf(':')
    if (i === -1) {
      continue
    }
    const typ = sys.system.slice(0, i)

    const typeSys = typeMap[typ]
    if (typeSys) {
      typeSys.rate += sys.rate
      continue
    }

    typeMap[typ] = {
      ...sys,
      system: typ,
    }
  }

  for (let sysType in typeMap) {
    const sys = typeMap[sysType]
    systems.push({
      ...sys,
      system: sys.system + ':all',
    })
  }

  systems = orderBy(systems, 'system')

  return systems
}
