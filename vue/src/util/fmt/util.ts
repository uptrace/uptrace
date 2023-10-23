import numbro from 'numbro'

export function trimMantissa(n: number, mantissa: number): string {
  if (typeof n !== 'number') {
    return '0'
  }
  return numbro(n).format({
    mantissa,
    trimMantissa: true,
  })
}

export function fixedMantissa(n: number | undefined, mantissa: number): string {
  if (typeof n === 'undefined') {
    return '0'
  }
  return numbro(n).format({ mantissa })
}

export function mantissa(n: number): number {
  n = Math.abs(n)
  for (let i = 12; i >= -1; i--) {
    const threshold = 1 / Math.pow(10, i)
    if (n < threshold) {
      return i + 2
    }
  }
  return 0
}
