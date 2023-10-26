const lerp = (a: number, b: number, t: number) => (b - a) * t + a
const unlerp = (a: number, b: number, t: number) => (t - a) / (b - a)

export function mapNumber(t: number, a1: number, b1: number, a2: number, b2: number) {
  if (isNaN(t)) {
    return a2
  }
  return lerp(a2, b2, unlerp(a1, b1, t))
}
