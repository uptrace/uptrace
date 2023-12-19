import numbro from 'numbro'

import { trimMantissa, mantissaSize, Config } from './util'

export function num(n: any, conf: Config = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 1000) {
    return trimMantissa(n, mantissaSize(n), conf)
  }

  for (let suffix of ['k', 'mln', 'bln', 'tln']) {
    n /= 1000
    abs /= 1000
    if (abs < 10) {
      return trimMantissa(n, 2, conf) + suffix
    }
    if (abs < 100) {
      return trimMantissa(n, 1, conf) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0, conf) + suffix
    }
  }

  return trimMantissa(n, 0, conf) + 'tln'
}

export function numShort(n: any, conf = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  let abs = Math.abs(n)

  if (abs < 0.001) {
    return round(n, mantissaSize(n)).toExponential()
  }
  if (abs < 1000) {
    return trimMantissa(n, mantissaSize(n), conf)
  }

  for (let suffix of ['k', 'mn', 'bn', 'tn']) {
    n /= 1000
    abs /= 1000
    if (abs < 100) {
      return trimMantissa(n, 1, conf) + suffix
    }
    if (abs < 1000) {
      return trimMantissa(n, 0, conf) + suffix
    }
  }

  return trimMantissa(n, 0, conf) + 'tn'
}

export function numVerbose(n: any, conf = {}): string {
  if (n === null) {
    return 'null'
  }
  if (typeof n !== 'number' || n === 0) {
    return '0'
  }

  return numbro(n).format({
    thousandSeparated: n >= 1e6,
    mantissa: mantissaSize(n),
    trimMantissa: true,
    ...conf,
  })
}

//------------------------------------------------------------------------------

export function percents(n: any, conf: Config = {}): string {
  if (n === null || isNaN(n)) {
    return 'null'
  }
  if (typeof n !== 'number') {
    return '0%'
  }
  const value = numbro(n).format({
    mantissa: mantissaSize(n),
    optionalMantissa: true,
    forceSign: conf.forceSign ?? false,
  })
  return value + '%'
}

export function utilization(n: any, conf: Config = {}): string {
  if (n === null || isNaN(n)) {
    return 'null'
  }
  if (typeof n !== 'number') {
    return '0%'
  }

  return numbro(n).format({
    output: 'percent',
    mantissa: mantissaSize(n * 100),
    optionalMantissa: true,
    forceSign: conf.forceSign ?? false,
  })
}

//------------------------------------------------------------------------------

function round(n: number, mantissa: number): number {
  const mul = Math.pow(10, mantissa)
  return Math.round((n + Number.EPSILON) * mul) / mul
}
