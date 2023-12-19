import { formatNum, trimMantissa, Config } from './util'

export function duration(n: number | undefined | null, conf: Config = {}): string {
  if (conf.short) {
    return durationShort(n, conf)
  }

  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 1) {
    return formatNum(n * 1000, 0, conf) + 'ns'
  }
  if (abs < 1000) {
    return formatNum(n, 0, conf) + 'µs'
  }

  n /= 1000
  abs = Math.abs(n)

  if (abs < 100) {
    return formatNum(n, 1, conf) + 'ms'
  }
  if (abs < 10000) {
    return formatNum(n, 0, conf) + 'ms'
  }

  n /= 1000
  abs = Math.abs(n)

  if (abs < 100) {
    return formatNum(n, 1, conf) + 's'
  }

  n /= 60
  abs = Math.abs(n)

  if (abs < 10) {
    return formatNum(n, 1, conf) + 'm'
  }
  if (abs < 100) {
    return formatNum(n, 0, conf) + 'm'
  }

  n /= 60
  abs = Math.abs(n)

  if (abs < 10) {
    return formatNum(n, 1, conf) + 'h'
  }
  return formatNum(n, 0, conf) + 'h'
}

export function durationShort(n: number | undefined | null, conf: Config = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 1) {
    return trimMantissa(n * 1000, 0, conf) + 'ns'
  }
  if (abs < 1000) {
    return trimMantissa(n, 0, conf) + 'µs'
  }

  n /= 1000
  abs = Math.abs(n)

  if (abs < 10) {
    return trimMantissa(n, 1, conf) + 'ms'
  }
  if (abs < 1000) {
    return trimMantissa(n, 0, conf) + 'ms'
  }

  n /= 1000
  abs = Math.abs(n)

  if (abs < 100) {
    return trimMantissa(n, 1, conf) + 's'
  }

  n /= 60
  abs = Math.abs(n)

  if (abs < 10) {
    return trimMantissa(n, 1, conf) + 'm'
  }
  if (abs < 100) {
    return trimMantissa(n, 0, conf) + 'm'
  }

  n /= 60
  abs = Math.abs(n)

  if (abs < 10) {
    return trimMantissa(n, 1, conf) + 'h'
  }
  return trimMantissa(n, 0, conf) + 'h'
}
