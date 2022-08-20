import { fixedMantissa, trimMantissa } from './util'

export function durationFixed(n: number | undefined, conf = {}): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0, conf) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, conf) + 'µs'
  }
  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0, conf) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, conf) + 'ms'
  }
  if (Math.abs(n) < 10000) {
    return fixedMantissa(n, 0, conf) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, conf) + 's'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, conf) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0, conf) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, conf) + 'h'
  }
  return fixedMantissa(n, 0, conf) + 'h'
}

export function duration(n: number | undefined, conf = {}): string {
  return durationFixed(n, { trimMantissa: true, ...conf })
}

export function durationShort(n: number | undefined, conf = {}): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, conf) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, conf) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 10) {
    return trimMantissa(n, 1, conf) + 'ms'
  }
  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, conf) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1, conf) + 's'
  }

  n /= 60

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1, conf) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0, conf) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, conf) + 'h'
  }
  return fixedMantissa(n, 0, conf) + 'h'
}
