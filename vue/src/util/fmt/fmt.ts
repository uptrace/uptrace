import { Unit } from '@/util/fmt/unit'
import { datetime } from '@/util/fmt/date'
import { num, numVerbose, numShort, percents, utilization } from '@/util/fmt/num'
import { bytes, bytesShort } from '@/util/fmt/bytes'
import { duration, durationShort } from '@/util/fmt/duration'

export interface Data extends Record<string, any> {
  time: string[]
  count?: number[]
  rate: number[]
  errorPct?: number[]
  p50?: number[]
  p90?: number[]
  p99?: number[]
  max?: number[]
  minTime?: string
  maxTime?: string
}

export function fmt(val: any, unit = '', ...args: any[]): string {
  return createFormatter(unit)(val, ...args)
}

export type Formatter = (val: any, ...args: any[]) => string

export function createFormatter(unit: string): Formatter {
  if (typeof unit === 'function') {
    return unit
  }

  switch (unit) {
    case Unit.None:
      return none
    case Unit.Percents:
      return percents
    case Unit.Utilization:
      return utilization
    case Unit.Time:
      return datetime
    case Unit.Nanoseconds:
      return duration
    case Unit.Microseconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e3), ...args)
      }
    case Unit.Milliseconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e6), ...args)
      }
    case Unit.Seconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e9), ...args)
      }
    case Unit.Bytes:
      return bytes
    default:
      return (val: any) => {
        val = none(val)
        if (unit) {
          val += ' ' + unwrapUnit(unit)
        }
        return val
      }
  }
}

function unwrapUnit(unit: string): string {
  if (unit.length <= 2) {
    return unit
  }
  if (unit.startsWith('{') && unit.endsWith('}')) {
    return unit.slice(1, unit.length - 1)
  }
  return unit
}

export function createShortFormatter(unit: Unit | string | Formatter): Formatter {
  if (typeof unit === 'function') {
    return unit
  }

  switch (unit) {
    case Unit.Percents:
      return percents
    case Unit.Utilization:
      return utilization
    case Unit.Time:
      return datetime
    case Unit.Nanoseconds:
      return durationShort
    case Unit.Microseconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 1e3), ...args)
      }
    case Unit.Milliseconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 1e6), ...args)
      }
    case Unit.Seconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 1e9), ...args)
      }
    case Unit.Bytes:
      return bytesShort
    default:
      return (val: any) => {
        // No unit.
        return noneShort(val)
      }
  }
}

export function fmtShort(val: any, unit: Unit | string, ...args: any[]): string {
  return createShortFormatter(unit)(val, ...args)
}

export function createVerboseFormatter(unit: Unit | string | Formatter): Formatter {
  if (typeof unit === 'function') {
    return unit
  }

  const fn = createFormatter(unit)
  if (fn === none) {
    return numVerbose
  }
  return fn
}

function none(v: unknown, conf = {}): string {
  if (v === undefined) {
    return '<undefined>'
  }
  if (v === '') {
    return '<empty string>'
  }
  if (typeof v === 'number') {
    return num(v, conf)
  }
  return String(v)
}

function noneShort(v: unknown): string {
  if (v === undefined) {
    return '<undefined>'
  }
  if (v === '') {
    return '<empty string>'
  }
  if (typeof v === 'number') {
    return numShort(v)
  }
  return String(v)
}

function adjustNumber(v: any, mod: number): any {
  if (typeof v === 'number') {
    return v * mod
  }
  return v
}
