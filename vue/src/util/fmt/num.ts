import numbro from 'numbro'

import { trimMantissa, mantissa } from './util'

export function num(n: number | undefined): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 1000) {
    return trimMantissa(n, mantissa(n))
  }

  for (let suffix of ['k', 'mln', 'bln', 'tln']) {
    n /= 1000
    abs /= 1000
    if (abs < 10) {
      return trimMantissa(n, 2) + suffix
    }
    if (abs < 100) {
      return trimMantissa(n, 1) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0) + suffix
    }
  }

  return trimMantissa(n, 0) + 'tln'
}

export function numShort(n: number | undefined, opts = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 0.001) {
    return round(n, mantissa(n)).toExponential()
  }
  if (abs < 1000) {
    return trimMantissa(n, mantissa(n))
  }

  for (let suffix of ['k', 'mn', 'bn', 'tn']) {
    n /= 1000
    abs /= 1000
    if (abs < 100) {
      return trimMantissa(n, 1) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0) + suffix
    }
  }

  return trimMantissa(n, 0) + 'tn'
}

export function numVerbose(n: any = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  return numbro(n).format({
    thousandSeparated: n >= 1e6,
    mantissa: mantissa(n),
    trimMantissa: true,
  })
}

//------------------------------------------------------------------------------

export function percents(n: any): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || Math.abs(n) < 0.0001) {
    return '0%'
  }
  const value = numbro(n).format({
    mantissa: percentsMantissa(n),
    optionalMantissa: true,
  })
  return value + '%'
}

export function utilization(n: any): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || Math.abs(n) < 0.0001) {
    return '0%'
  }

  return numbro(n).format({
    output: 'percent',
    mantissa: percentsMantissa(n),
    optionalMantissa: true,
  })
}

function percentsMantissa(n: number): number {
  n = Math.abs(n)
  if (n < 0.01) {
    return 2
  }
  if (n < 0.1) {
    return 1
  }
  return 0
}

//------------------------------------------------------------------------------

export function bytes(n: number | undefined): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)
  if (abs < 10000) {
    return trimMantissa(n, 0) + 'by'
  }

  for (let suffix of ['KB', 'MB', 'GB', 'TB', 'PB']) {
    n /= 1000
    abs /= 1000
    if (abs < 10) {
      return trimMantissa(n, 2) + suffix
    }
    if (abs < 100) {
      return trimMantissa(n, 1) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0) + suffix
    }
  }

  return trimMantissa(n, 0) + 'PB'
}

export function bytesShort(n: number | undefined, opts = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)
  if (abs < 100) {
    return trimMantissa(n, 0) + 'by'
  }

  for (let suffix of ['KB', 'MB', 'GB', 'TB', 'PB']) {
    n /= 1000
    abs /= 1000
    if (abs < 100) {
      return trimMantissa(n, 1) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0) + suffix
    }
  }

  return trimMantissa(n, 0) + 'PB'
}

function round(n: number, mantissa: number): number {
  const mul = Math.pow(10, mantissa)
  return Math.round((n + Number.EPSILON) * mul) / mul
}
