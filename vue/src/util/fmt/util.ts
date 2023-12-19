import numbro from 'numbro'

export interface Config {
  trimMantissa?: boolean
  forceSign?: boolean
  short?: boolean
}

export function fixedMantissa(n: unknown, mantissa: number, conf: Config = {}): string {
  return formatNum(n, mantissa, { ...conf, trimMantissa: false })
}

export function trimMantissa(n: unknown, mantissa: number, conf: Config = {}): string {
  return formatNum(n, mantissa, { ...conf, trimMantissa: true })
}

export function formatNum(n: unknown, mantissa: number, conf: Config = {}): string {
  if (typeof n !== 'number') {
    return '0'
  }
  return numbro(n).format({
    mantissa,
    trimMantissa: conf.trimMantissa ?? true,
    forceSign: conf.forceSign ?? false,
  })
}

export function mantissaSize(n: number): number {
  n = Math.abs(n)
  for (let i = 12; i >= -1; i--) {
    const threshold = 1 / Math.pow(10, i)
    if (n < threshold) {
      return i + 2
    }
  }
  return 0
}

export function inflect(n: number, singular: string, plural: string): string {
  if (n === 1) {
    return singular
  }
  return plural
}
