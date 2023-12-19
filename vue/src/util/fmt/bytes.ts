import { formatNum, trimMantissa, Config } from './util'

export function bytes(n: any, conf: Config = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)
  if (abs < 10000) {
    return formatNum(n, 0, conf) + 'by'
  }

  for (let suffix of ['KB', 'MB', 'GB', 'TB', 'PB']) {
    n /= 1000
    abs /= 1000
    if (abs < 10) {
      return formatNum(n, 2, conf) + suffix
    }
    if (abs < 100) {
      return formatNum(n, 1, conf) + suffix
    }
    if (abs < 1000) {
      return formatNum(n, 0, conf) + suffix
    }
  }

  return formatNum(n, 0, conf) + 'PB'
}

export function bytesShort(n: any, conf: Config = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)
  if (abs < 100) {
    return trimMantissa(n, 0, conf) + 'by'
  }

  for (let suffix of ['KB', 'MB', 'GB', 'TB', 'PB']) {
    n /= 1000
    abs /= 1000
    if (abs < 100) {
      return trimMantissa(n, 1, conf) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0, conf) + suffix
    }
  }

  return trimMantissa(n, 0, conf) + 'PB'
}
