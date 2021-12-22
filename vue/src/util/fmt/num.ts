import numbro from 'numbro'

import { fixedMantissa, trimMantissa } from './util'

export function num(n: number | undefined): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }
  if (Math.abs(n) < 1) {
    return trimMantissa(n, 3)
  }
  if (Math.abs(n) < 1000) {
    return trimMantissa(n, 1)
  }
  n /= 1000
  if (Math.abs(n) < 100) {
    return fixedMantissa(n, 1) + 'k'
  }
  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 0) + 'k'
  }
  n /= 1000
  if (Math.abs(n) < 1000) {
    return fixedMantissa(n, 1) + 'm'
  }
  n /= 1000
  return fixedMantissa(n, 1) + 'b'
}

export function percent(n: any): string {
  if (typeof n !== 'number' || Math.abs(n) < 0.0001) {
    return '0%'
  }

  return numbro(n).format({
    output: 'percent',
    mantissa: percentMantissa(n),
    optionalMantissa: true,
  })
}

function percentMantissa(n: number): number {
  n = Math.abs(n)
  if (n < 0.01) {
    return 2
  }
  if (n < 0.1) {
    return 1
  }
  return 0
}

export function bytes(n: number | undefined): string {
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  if (Math.abs(n) < 10) {
    return trimMantissa(n, 2)
  }
  if (Math.abs(n) < 1024) {
    return trimMantissa(n, 1)
  }

  for (let unit of ['KB', 'MB', 'GB', 'TB', 'PB']) {
    n /= 1024

    if (Math.abs(n) < 10) {
      return trimMantissa(n, 1) + unit
    }
    n = Math.round(n) // round before compare
    if (Math.abs(n) < 1000) {
      return trimMantissa(n, 0) + unit
    }
  }

  return trimMantissa(n, 0) + 'PB'
}
