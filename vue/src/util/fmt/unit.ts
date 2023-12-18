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

  Date = '{date}',
  Time = 'time',
  UnixTime = 'unix-time', // number of seconds
}

export const UNITS = [
  Unit.None,
  Unit.Bytes,
  Unit.Nanoseconds,
  Unit.Microseconds,
  Unit.Milliseconds,
  Unit.Seconds,
  Unit.Percents,
  Unit.Utilization,
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
