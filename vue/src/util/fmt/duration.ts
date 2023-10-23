import { fixedMantissa, trimMantissa } from './util'

export function duration(n: number | undefined): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1) + 'µs'
  }
  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1) + 'ms'
  }
  if (Math.abs(n) < 10000) {
    return fixedMantissa(n, 0) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1) + 's'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1) + 'h'
  }
  return fixedMantissa(n, 0) + 'h'
}

export function durationShort(n: number | undefined): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0) + 'ns'
  }

  n /= 1000

  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0) + 'µs'
  }

  n /= 1000

  if (Math.abs(n) < 10) {
    return trimMantissa(n, 1) + 'ms'
  }
  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 0) + 'ms'
  }

  n /= 1000

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1) + 's'
  }

  n /= 60

  if (Math.abs(n) < 100) {
    return trimMantissa(n, 1) + 'm'
  }
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 0) + 'm'
  }

  n /= 60

  if (Math.abs(n) < 10) {
    return fixedMantissa(n, 1) + 'h'
  }
  return fixedMantissa(n, 0) + 'h'
}
