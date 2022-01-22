import { fixedMantissa, trimMantissa } from './util'

export function durationFixed(n: number | undefined, cfg = {}): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0, cfg) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, cfg) + 'µs'
  }
  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0, cfg) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, cfg) + 'ms'
  }
  if (Math.abs(n) < 10000) {
    return fixedMantissa(n, 0, cfg) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1, cfg) + 's'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, cfg) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0, cfg) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, cfg) + 'h'
  }
  return fixedMantissa(n, 0, cfg) + 'h'
}

export function durationShort(n: number | undefined, cfg = {}): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, cfg) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, cfg) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 10) {
    return trimMantissa(n, 1, cfg) + 'ms'
  }
  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0, cfg) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1, cfg) + 's'
  }

  n /= 60

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1, cfg) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0, cfg) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1, cfg) + 'h'
  }
  return fixedMantissa(n, 0, cfg) + 'h'
}
