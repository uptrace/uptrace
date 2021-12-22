import { Unit } from './unit'
import { num, bytes, percent } from './num'
import { durationShort } from './duration'

export * from './unit'
export * from './duration'
export * from './num'

export type Formatter = (val: any, ...args: any[]) => string

export function createFormatter(unit: string | Unit): Formatter {
  switch (unit) {
    case Unit.Percents:
      return percent
    case Unit.Nanoseconds:
      return durationShort
    case Unit.Bytes:
      return bytes
    default:
      return none
  }
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
