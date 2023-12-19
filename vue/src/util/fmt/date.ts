import enGB from 'date-fns/locale/en-GB'
import {
  isValid,
  parse,
  parseISO,
  format,
  formatDistanceToNow,
  formatRelative,
  addMilliseconds,
  subMilliseconds,
} from 'date-fns'

export const MILLISECOND = 1
export const SECOND = 1000 * MILLISECOND
export const MINUTE = 60 * SECOND
export const HOUR = 60 * MINUTE
export const DAY = 24 * HOUR

export function toDate(v: any): Date {
  if (v instanceof Date) {
    return v
  }
  if (typeof v === 'number') {
    return new Date(v)
  }
  if (typeof v !== 'string') {
    return new Date(NaN)
  }

  if (/^\d+$/.test(v)) {
    return new Date(parseInt(v, 10) / 1e6)
  }

  const unix = Date.parse(v)
  return new Date(unix)
}

export function date(v: number | string | Date | undefined, fmt = 'LLL d y'): string {
  return formatDate(v, fmt)
}

export function dateShort(v: number | string | Date | undefined, fmt = 'LLL d'): string {
  return formatDate(v, fmt)
}

export function datetime(v: number | string | Date | undefined, fmt = 'LLL d y HH:mm:ss'): string {
  return formatDate(v, fmt)
}

export function datetimeShort(v: number | string | Date | undefined, fmt = 'LLL d HH:mm'): string {
  return formatDate(v, fmt)
}

export function datetimeFull(
  v: number | string | Date | undefined,
  fmt = 'LLL d y HH:mm:ss.SSS',
): string {
  return formatDate(v, fmt)
}

export function time(v: number | string | Date | undefined, fmt = 'HH:mm:ss.SSS') {
  return formatDate(v, fmt)
}

function formatDate(v: number | string | Date | undefined, fmt: string): string {
  if (!v) {
    return String(v)
  }

  if (typeof v === 'string') {
    const date = parseISO(v)
    if (!isNaN(date.getTime())) {
      return format(date, fmt)
    }
    return String(v)
  }

  return format(v, fmt)
}

const formatRelativeLocale = {
  lastWeek: "'last' eeee 'at' p",
  yesterday: "'yesterday at' p",
  today: "'today at' p",
  tomorrow: "'tomorrow at' p",
  nextWeek: "eeee 'at' p",
  other: 'P',
}

const locale = {
  ...enGB,
  formatRelative: (token: keyof typeof formatRelativeLocale) => formatRelativeLocale[token],
}

export function relative(v: number | string | Date | undefined): string {
  if (!v) {
    return String(v)
  }
  if (typeof v === 'string') {
    v = parseISO(v)
  }
  return formatRelative(v, new Date(), { locale })
}

export function fromNow(v: number | string | Date): string {
  return formatDistanceToNow(toDate(v))
}

const basicFormat = "yyyyMMdd'T'HHmmss"

export function formatUTC(dt: Date): string {
  return format(toUTC(dt), basicFormat)
}

export function parseUTC(s: string): Date {
  let dt = parse(s, basicFormat, new Date())
  if (isValid(dt)) {
    dt = toLocal(dt)
    return dt
  }
  return new Date(s)
}

export function toUTC(dt: Date): Date {
  return addMilliseconds(dt, dt.getTimezoneOffset() * MINUTE)
}

export function toLocal(dt: Date): Date {
  return subMilliseconds(dt, dt.getTimezoneOffset() * MINUTE)
}

export function ceilDate(dt: Date, prec: number): Date {
  let r = dt.getTime() % prec
  if (r === 0) {
    return dt
  }
  return addMilliseconds(dt, prec - r)
}

export function truncDate(dt: Date, prec: number): Date {
  let r = dt.getTime() % prec
  if (r === 0) {
    return dt
  }
  return subMilliseconds(dt, r)
}
