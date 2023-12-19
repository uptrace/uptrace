import Color from 'color'

// Misc
import { AttrKey, SystemName } from '@/models/otel'
import { walkTree, Tree } from '@/models/tree'
import { eChart as colorSet } from '@/util/colorscheme'

const SYSTEM_LIMIT = 6

export interface ColoredSystem {
  name: string
  duration: number
  color: string
  barStyle?: { [key: string]: string }
}

type ColoredSystemMap = Record<string, ColoredSystem>

export interface ColoredSpan<T> extends Tree<T> {
  system: string
  durationSelf: number
  attrs: Record<string, any>

  _systemName: string
  color: SpanColor
}

export interface SpanColor {
  base: string
  lighten: string
  darken: string
}

export function spanColoredSystems<T extends ColoredSpan<T>>(root: T): ColoredSystem[] {
  const sysMap: ColoredSystemMap = {}

  walkTree(root, (span, parent) => {
    span._systemName = span.system
    if (span._systemName === SystemName.Funcs) {
      const service = parent?._systemName ?? span.attrs[AttrKey.serviceName]
      if (service) {
        span._systemName = service
      }
    }

    if (!span._systemName) {
      return true
    }

    let sysInfo = sysMap[span._systemName]
    if (!sysInfo) {
      sysInfo = {
        name: span._systemName,
        duration: 0,
        color: '',
      }
      sysMap[span._systemName] = sysInfo
    }

    const dur = span.durationSelf
    if (dur) {
      sysInfo.duration += dur
    }

    return true
  })

  let systems = systemMapToList(sysMap)
  const otherSystem = {
    name: 'other',
    duration: 0,
    color: colorSet[colorSet.length - 1],
  }

  for (let system of systems.slice(SYSTEM_LIMIT)) {
    otherSystem.duration += system.duration
    delete sysMap[system.name]
  }

  systems = systems.slice(0, SYSTEM_LIMIT)
  if (otherSystem.duration > 0) {
    systems.push(otherSystem)
  }

  // last color is reserved for other
  for (let i = 0; i < systems.length && i < colorSet.length - 1; i++) {
    const system = systems[i]
    system.color = colorSet[i]
  }

  walkTree(root, (span) => {
    const color = sysMap[span._systemName]?.color || otherSystem.color
    span.color = {
      base: color,
      lighten: Color(color).lighten(0.33).hex(),
      darken: Color(color).darken(0.33).hex(),
    }
    return true
  })

  return systems
}

function systemMapToList(map: ColoredSystemMap): ColoredSystem[] {
  const list: ColoredSystem[] = []

  for (let k in map) {
    list.push(map[k])
  }

  list.sort((a, b) => {
    if (a.duration < b.duration) return 1
    if (a.duration > b.duration) return -1
    return 0
  })

  for (let i = 0; i < list.length && i < colorSet.length - 1; i++) {
    const system = list[i]
    system.color = colorSet[i]
  }

  return list
}
