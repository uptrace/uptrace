import { AttrKey } from '@/models/otel'

export enum Metric {
  Count = 'count',
  Rate = 'rate',
  ErrorPct = 'errorPct',
  Failures = 'failures',
  Duration = 'duration',
  P50 = 'p50',
  P90 = 'p90',
  P99 = 'p99',
  Max = 'max',
}

export enum Unit {
  None = '',
  Percents = 'percents',
  Utilization = 'utilization',

  Nanoseconds = 'nanoseconds',
  Microseconds = 'microseconds',
  Milliseconds = 'milliseconds',
  Seconds = 'seconds',

  Bytes = 'bytes',
  Kilobytes = 'kilobytes',
  Megabytes = 'megabytes',
  Gigabytes = 'gigabytes',
  Terabytes = 'terabytes',

  Date = 'date',
  Time = 'time',
  UnixTime = 'unix-time', // in seconds
}

export const UNITS = [
  Unit.None,
  Unit.Bytes,
  Unit.Nanoseconds,
  Unit.Microseconds,
  Unit.Milliseconds,
  Unit.Seconds,
  Unit.Percents,
]

export function isCustomUnit(unit: string): boolean {
  return unit.startsWith('{') && unit.endsWith('}')
}

export function unitShortName(unit: Unit): string {
  switch (unit) {
    case Unit.None:
      return ''
    case Unit.Percents:
      return '%'
    case Unit.Utilization:
      return 'util'

    case Unit.Nanoseconds:
      return 'ns'
    case Unit.Microseconds:
      return 'us'
    case Unit.Milliseconds:
      return 'ms'
    case Unit.Seconds:
      return 's'

    case Unit.Bytes:
      return 'by'
    case Unit.Kilobytes:
      return 'kb'
    case Unit.Megabytes:
      return 'mb'
    case Unit.Gigabytes:
      return 'gb'
    case Unit.Terabytes:
      return 'tb'

    default:
      return unit
  }
}

export function unitFromName(name: Metric | string, value?: unknown): Unit {
  const isNum = typeof value === 'number'

  if (!isNum && value !== undefined) {
    return Unit.None
  }

  switch (name) {
    case Metric.Count:
    case Metric.Rate:
      return Unit.None
    case Metric.ErrorPct:
    case Metric.Failures:
      return Unit.Percents
    case Metric.Duration:
    case Metric.P50:
    case Metric.P90:
    case Metric.P99:
    case Metric.Max:
      return Unit.Microseconds
  }

  let key = ''

  const m = name.match(/(\S+)\((\S+)\)/)
  if (m) {
    key = m[2]
  } else {
    key = name
  }

  if (isDurationField(key)) {
    return Unit.Microseconds
  }
  if (isByteField(key)) {
    return Unit.Bytes
  }
  if (isDateField(key)) {
    return Unit.Time
  }
  if (isPercentField(key)) {
    return Unit.Percents
  }

  return Unit.None
}

export function isDurationField(s: string): boolean {
  return s === AttrKey.spanDuration || hasPart(s, 'duration') || hasPart(s, 'dur')
}

export function isByteField(s: string): boolean {
  return hasPart(s, 'bytes')
}

export function isDateField(s: string): boolean {
  return s === AttrKey.spanTime || hasPart(s, 'time') || hasPart(s, 'date')
}

export function isRateField(s: string): boolean {
  return s === AttrKey.spanCountPerMin || hasPart(s, 'rate')
}

export function isCountField(s: string): boolean {
  return s === AttrKey.spanCount || hasPart(s, 'count')
}

export function isPercentField(s: string): boolean {
  return s === AttrKey.spanErrorRate || hasPart(s, 'pct') || hasPart(s, 'percent')
}

function hasPart(str: string, substr: string): boolean {
  return (
    str === substr || re(`^${substr}[^a-zA-Z]`).test(str) || re(`[^a-zA-Z]${substr}$`).test(str)
  )
}

function re(s: string) {
  return new RegExp(s, 'i')
}
