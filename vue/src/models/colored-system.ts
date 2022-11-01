import colors, { Color } from 'vuetify/lib/util/colors'

// Utilities
import { AttrKey, SystemName } from '@/models/otel'
import { walkTree, Tree } from '@/models/tree'

interface SystemColor {
  lighten: string
  base: string
  darken: string
}

const colorSet = ['blue', 'pink', 'green', 'orange', 'purple', 'teal'].map((name) => {
  const color = colors[name as keyof typeof colors] as Color
  return {
    lighten: color.lighten4,
    base: color.base,
    darken: color.darken2,
  }
})

export interface ColoredSystem {
  system: string
  duration: number
  color: SystemColor
  barStyle?: { [key: string]: string }
}

type ColoredSystemMap = Record<string, ColoredSystem>

export interface ColoredSpan<T> extends Tree<T> {
  system: string
  durationSelf: number
  attrs: Record<string, any>

  _system: string
  color: string
  lightenColor: string
  darkenColor: string
}

export function spanColoredSystems<T extends ColoredSpan<T>>(root: T): ColoredSystem[] {
  const sysMap: ColoredSystemMap = {}

  walkTree(root, (span, parent) => {
    span._system = span.system
    if (span._system === SystemName.funcs) {
      const service = parent?._system ?? span.attrs[AttrKey.serviceName]
      if (service) {
        span._system = service
      }
    }

    if (!span._system) {
      return true
    }

    let sysInfo = sysMap[span._system]
    if (!sysInfo) {
      sysInfo = {
        system: span._system,
        duration: 0,
        color: undefined as unknown as SystemColor,
      }
      sysMap[span._system] = sysInfo
    }

    const dur = span.durationSelf
    if (dur) {
      sysInfo.duration += dur
    }

    return true
  })

  const systems = systemMapToList(sysMap)
  const otherColor = colorSet[colorSet.length - 1]

  walkTree(root, (span) => {
    const color = sysMap[span._system]?.color ?? otherColor
    span.color = color.base
    span.lightenColor = color.lighten
    span.darkenColor = color.darken
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
