import { Unit } from './unit'
import { datetime } from '@/util/fmt/date'
import { num, numShort, bytes, bytesShort, percents, utilization } from './num'
import { duration, durationShort } from './duration'

export * from './unit'
export * from './duration'
export * from './num'

export function fmt(val: any, unit = '', ...args: any[]): string {
  return createFormatter(unit)(val, ...args)
}

export type Formatter = (val: any, ...args: any[]) => string

export function createFormatter(unit: string | Formatter): Formatter {
  if (typeof unit === 'function') {
    return unit
  }

  switch (unit) {
    case Unit.Percents:
      return percents
    case Unit.Utilization:
      return utilization
    case Unit.Date:
      return datetime
    case Unit.Nanoseconds:
      return duration
    case Unit.Microseconds:
      return (val: any) => {
        return duration(adjustNumber(val, 1e3))
      }
    case Unit.Milliseconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e6))
      }
    case Unit.Seconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e9))
      }
    case Unit.Bytes:
      return bytes
  }
  return none
}

export function createShortFormatter(unit = ''): Formatter {
  switch (unit) {
    case Unit.Percents:
      return percents
    case Unit.Utilization:
      return utilization
    case Unit.Date:
      return datetime
    case Unit.Nanoseconds:
      return durationShort
    case Unit.Microseconds:
      return (val: any) => {
        return durationShort(adjustNumber(val, 1e3))
      }
    case Unit.Milliseconds:
      return (val: any) => {
        return durationShort(adjustNumber(val, 1e6))
      }
    case Unit.Seconds:
      return (val: any) => {
        return durationShort(adjustNumber(val, 1e9))
      }
    case Unit.Bytes:
      return bytesShort
  }
  return noneShort
}

function none(v: unknown): string {
  if (v === undefined) {
    return '<missing>'
  }
  if (v === '') {
    return '<empty>'
  }
  if (typeof v === 'number') {
    return num(v)
  }
  return String(v)
}

function noneShort(v: unknown): string {
  if (v === undefined) {
    return '<missing>'
  }
  if (v === '') {
    return '<empty>'
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
