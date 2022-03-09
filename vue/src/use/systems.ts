import { clone, orderBy } from 'lodash'
import { shallowRef, computed, proxyRefs } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

export interface System {
  system: string
  isEvent: boolean

  count: number
  countPerMin: number
  errorCount: number

  dummy?: boolean
  numChild?: number
}

export interface SystemTree extends System {
  children?: SystemTree[]
}

export type UseSystems = ReturnType<typeof useSystems>

export function useSystems(dateRange: UseDateRange) {
  const { route } = useRouter()

  const { loading, data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/tracing/${projectId}/systems`,
      params: {
        ...dateRange.axiosParams(),
      },
    }
  })

  const systems = computed((): System[] => {
    return addDummySystems(data.value?.systems ?? [])
  })

  const tree = computed((): SystemTree[] => {
    return buildSystemTree(systems.value as SystemTree[])
  })

  const hasNoData = computed(() => {
    if (Array.isArray(data.value?.systems)) {
      return data.value.systems.length === 0
    }
    return false
  })

  const internalValue = shallowRef<string>()

  const systemValue = computed({
    get() {
      return internalValue.value
    },
    set(system: string | undefined) {
      internalValue.value = system ? system : undefined
    },
  })

  const system = computed((): System | undefined => {
    if (!systemValue.value) {
      return undefined
    }

    const system = systems.value.find((sys) => sys.system === systemValue.value)
    return system
  })

  const isEvent = computed((): boolean => {
    return system.value?.isEvent ?? false
  })

  function axiosParams() {
    return {
      system: system.value?.system,
    }
  }

  function change(system: string): void {
    systemValue.value = system
  }

  function reset(): void {
    systemValue.value = undefined
  }

  return proxyRefs({
    loading,

    list: systems,
    tree,
    hasNoData,

    activeValue: systemValue,
    activeItem: system,
    isEvent,

    axiosParams,
    change,
    reset,
  })
}

function addDummySystems(systems: System[]): System[] {
  if (!systems.length) {
    return []
  }
  systems = clone(systems)

  const typeMap: Record<string, SystemTree> = {}

  for (let sys of systems) {
    const i = sys.system.indexOf(':')
    if (i === -1) {
      continue
    }
    const typ = sys.system.slice(0, i)

    let typeSys = typeMap[typ]
    if (!typeSys) {
      typeSys = {
        system: typ,
        isEvent: false,
        count: 0,
        countPerMin: 0,
        errorCount: 0,
        dummy: true,
        numChild: 0,
      }
      typeMap[typ] = typeSys
    }

    typeSys.count += sys.count
    typeSys.countPerMin += sys.countPerMin
    typeSys.errorCount += sys.errorCount
    typeSys.numChild!++
  }

  for (let sysType in typeMap) {
    const sys = typeMap[sysType]
    systems.push({
      ...sys,
      system: sys.system + ':all',
    })
  }

  systems = orderBy(systems, 'system')

  const internalIndex = systems.findIndex((sys) => sys.system === 'internal')
  if (internalIndex >= 0) {
    const internal = systems[internalIndex]
    internal.dummy = true
    systems.splice(internalIndex, 1)
    systems.unshift(internal)
  }

  systems.unshift({
    system: 'all',
    isEvent: false,
    count: 0,
    countPerMin: 0,
    errorCount: 0,

    dummy: true,
    numChild: 0,
  })

  return systems
}

function buildSystemTree(systems: SystemTree[]): SystemTree[] {
  systems = clone(systems)
  systems = systems.filter((sys) => sys.numChild !== 1)

  systems.slice(0).forEach((sys) => {
    if (!sys.system.endsWith(':all')) {
      return
    }

    const children = []

    const prefix = sys.system.slice(0, -'all'.length)
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

    sys.children = children
  })

  return systems
}
