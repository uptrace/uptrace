import numbro from 'numbro'

export function trimMantissa(n: number, mantissa: number, cfg = {}): string {
  return fixedMantissa(n, mantissa, {
    trimMantissa: true,
    ...cfg,
  })
}

export function fixedMantissa(n: number | undefined, mantissa: number, cfg = {}): string {
  if (typeof n === 'undefined') {
    return '0'
  }
  return numbro(n).format({ mantissa, ...cfg })
}
