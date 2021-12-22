import { xkey } from '@/models/otelattr'

export enum Unit {
  None = '',
  Bytes = 'bytes',
  Nanoseconds = 'nanoseconds',
  Percents = 'percents',
}

export function unitFromName(name: string, value?: unknown): Unit {
  const isNum = typeof value === 'number'

  if (!isNum && value !== undefined) {
    return Unit.None
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
  return s === xkey.spanDuration || hasField(s, 'duration')
}

export function isByteField(s: string): boolean {
  return hasField(s, 'bytes')
}

export function isPercentField(s: string): boolean {
  return s === xkey.spanErrorPct || hasField(s, 'pct')
}

function hasField(s: string, field: string): boolean {
  return s.endsWith('.' + field) || s.endsWith('_' + field)
}
