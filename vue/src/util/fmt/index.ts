import { Unit } from './unit'
import { datetime } from '@/util/fmt/date'
import { num, numShort, bytes, bytesShort, percent } from './num'
import { duration, durationShort } from './duration'

export * from './unit'
export * from './duration'
export * from './num'

export type Formatter = (val: any, ...args: any[]) => string

export function createFormatter(unit: string | Unit | Formatter): Formatter {
  if (typeof unit === 'function') {
    return unit
  }

  switch (unit) {
    case Unit.Percents:
      return percent
    case Unit.Date:
      return datetime
    case Unit.Nanoseconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 0.001), ...args)
      }
    case Unit.Microseconds:
      return duration
    case Unit.Milliseconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e3), ...args)
      }
    case Unit.Seconds:
      return (val: any, ...args: any[]) => {
        return duration(adjustNumber(val, 1e6), ...args)
      }
    case Unit.Bytes:
      return bytes
  }
  return none
}

export function createShortFormatter(unit: Unit | string): Formatter {
  switch (unit) {
    case Unit.Percents:
      return percent
    case Unit.Date:
      return datetime
    case Unit.Nanoseconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 0.001), ...args)
      }
    case Unit.Microseconds:
      return durationShort
    case Unit.Milliseconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 1e3), ...args)
      }
    case Unit.Seconds:
      return (val: any, ...args: any[]) => {
        return durationShort(adjustNumber(val, 1e6), ...args)
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
