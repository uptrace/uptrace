import { AttrKey } from '@/models/otel'

export enum Unit {
  None = '',
  Percents = 'percents',

  Nanoseconds = 'nanoseconds',
  Microseconds = 'microseconds',
  Milliseconds = 'milliseconds',
  Seconds = 'seconds',

  Bytes = 'bytes',
  Kilobytes = 'kilobytes',
  Megabytes = 'megabytes',
  Gigabytes = 'gigabytes',
  Terabytes = 'terabytes',

  Date = '{date}',
}

export function unitShortName(unit: Unit): string {
  switch (unit) {
    case Unit.None:
      return ''
    case Unit.Percents:
      return '%'

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

export function unitFromName(name: string, value?: unknown): Unit {
  const isNum = typeof value === 'number'

  if (!isNum && value !== undefined) {
    return Unit.None
  }

  switch (name) {
    case 'count':
    case 'rate':
      return Unit.None
    case 'errorPct':
      return Unit.Percents
    case 'p50':
    case 'p90':
    case 'p99':
      return Unit.Nanoseconds
  }

  let key = ''

  const m = name.match(/(\S+)\((\S+)\)/)
  if (m) {
    key = m[2]
  } else {
    key = name
  }

  if (isDurationField(key)) {
    return Unit.Nanoseconds
  }
  if (isByteField(key)) {
    return Unit.Bytes
  }
  if (isPercentField(key)) {
    return Unit.Percents
  }

  return Unit.None
}

export function isDurationField(s: string): boolean {
  return s === AttrKey.spanDuration || hasField(s, 'duration')
}

export function isByteField(s: string): boolean {
  return hasField(s, 'bytes')
}

export function isPercentField(s: string): boolean {
  return s === AttrKey.spanErrorPct || hasField(s, 'pct')
}

function hasField(s: string, field: string): boolean {
  return s.endsWith('.' + field) || s.endsWith('_' + field)
}
