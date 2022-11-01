import { cloneDeep, orderBy } from 'lodash-es'
import { shallowRef, computed, proxyRefs } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

// Utilities
import { SystemName, isEventSystem } from '@/models/otelattr'

export interface System {
  projectId: number
  system: string
  text: string
  isEvent: boolean

  count: number
  rate: number
  errorCount: number
  errorPct: number

  dummy?: boolean
  numChildren?: number
}

export type SystemsFilter = (systems: System[]) => System[]

export interface SystemTreeNode extends System {
  children?: SystemTreeNode[]
}

export type UseSystems = ReturnType<typeof useSystems>

export function useSystems(params: () => Record<string, any>) {
  const { route } = useRouter()

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/systems`,
      params: params(),
    }
  })

  const systems = computed((): System[] => {
    const systems = data.value?.systems ?? []
    systems.forEach((item: System) => (item.text = item.system))

    return addDummySystems(systems ?? [])
  })

  const hasNoData = computed(() => {
    if (Array.isArray(data.value?.systems)) {
      return data.value.systems.length === 0
    }
    return false
  })

  const internalValue = shallowRef<string>()

  const activeSystem = computed({
    get() {
      return internalValue.value
    },
    set(system: string | undefined) {
      internalValue.value = system ? system : undefined
    },
  })

  const isEvent = computed((): boolean => {
    return isEventSystem(activeSystem.value)
  })

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

  function change(system: string): void {
    activeSystem.value = system
  }

  function reset(): void {
    activeSystem.value = undefined
  }

  return proxyRefs({
    loading,

    items: systems,
    hasNoData,

    activeSystem,
    isEvent,

    axiosParams,
    queryParams,
    change,
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
      typeSys.numChildren!++
      continue
    }

    typeMap[typ] = {
      ...sys,
      system: typ,
      dummy: true,
      numChildren: 1,
    }
  }

  for (let sysType in typeMap) {
    const sys = typeMap[sysType]
    systems.push({
      ...sys,
      system: sys.system + ':all',
      text: sys.system + ':all',
    })
  }

  systems = orderBy(systems, 'system')

  return systems
}

export function buildSystemsTree(rawSystems: System[]): SystemTreeNode[] {
  let systems = cloneDeep(rawSystems) as SystemTreeNode[]
  systems = systems.filter((sys) => sys.numChildren !== 1)

  systems.slice(0).forEach((sys) => {
    if (!sys.system.endsWith(':all')) {
      return
    }

    const children = []

    const prefix = sys.system.slice(0, -SystemName.all.length)
    for (let j = systems.length - 1; j >= 0; j--) {
      const child = systems[j]
      if (child.system == sys.system) {
        continue
      }

      if (child.system.startsWith(prefix)) {
        systems.splice(j, 1)
        children.push(child)
      }
    }

    if (children.length) {
      sys.children = children
    }
  })

  return systems
}
