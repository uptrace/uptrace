import { cloneDeep, orderBy } from 'lodash-es'
import { shallowRef, computed, proxyRefs } from 'vue'

// Composables
import { useRoute, Values } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

// Utilities
import { isEventSystem, isGroupSystem, SystemName } from '@/models/otel'
import { DataHint } from '@/org/types'

export interface System {
  system: string

  count: number
  rate: number
  errorCount: number
  errorRate: number
  groupCount: number

  isGroup?: boolean
  indent?: boolean
}

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
      url: `/internal/v1/tracing/${projectId}/systems`,
      params: params(),
    }
  })

  const systems = computed((): System[] => {
    let systems = data.value?.systems ?? []

    systems = cloneDeep(systems)
    addAllSystem(systems, SystemName.All)
    addAllSystemForEachType(systems)
    return orderBy(systems, 'system')
  })

  const dataHint = computed((): DataHint | undefined => {
    return data.value?.dataHint
  })

  const internalValue = shallowRef<string[]>([])

  const activeSystems = computed({
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

  const isEvent = computed(() => {
    return isEventSystem(...activeSystems.value)
  })

  function reset(): void {
    activeSystems.value = []
  }

  function axiosParams() {
    return {
      system: activeSystems.value,
    }
  }

  function queryParams() {
    return { system: activeSystems.value }
  }

  function parseQueryParams(queryParams: Values) {
    activeSystems.value = queryParams.array('system')
  }

  return proxyRefs({
    loading,

    items: systems,
    dataHint,

    activeSystems,
    isEvent,

    reset,
    axiosParams,

    queryParams,
    parseQueryParams,
  })
}

function addAllSystemForEachType(systems: System[]) {
  if (!systems.length) {
    return
  }

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
      typeSys.errorCount += sys.errorCount
      typeSys.errorRate += sys.errorRate
      typeSys.groupCount += sys.groupCount
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
      isGroup: true,
    })
  }
}

export function addAllSystem(systems: System[], systemName: string) {
  if (!systems.length) {
    return
  }

  const index = systems.findIndex((item) => item.system === systemName)
  if (index >= 0) {
    return
  }

  const system = createAllSystem(systems, systemName)
  systems.unshift(system)
}

function createAllSystem(systems: System[], systemName: string): System {
  const allSystem = {
    system: systemName,
    count: 0,
    rate: 0,
    errorCount: 0,
    errorRate: 0,
    groupCount: 0,
    isGroup: true,
  }
  for (let system of systems) {
    if (!isGroupSystem(system.system)) {
      allSystem.count += system.count
      allSystem.rate += system.rate
      allSystem.errorCount += system.errorCount
      allSystem.groupCount += system.groupCount
    }
  }
  allSystem.errorRate = allSystem.errorCount / allSystem.count
  return allSystem
}

export function systemTypes(systems: System[]): string[] {
  const types = new Set<string>()

  for (let sys of systems) {
    let systemType = sys.system

    const i = systemType.indexOf(':')
    if (i >= 0) {
      systemType = systemType.slice(0, i)
    }

    types.add(systemType)
  }

  return Array.from(types)
}
