export function parseJson(s: string): any {
  if (!isJson(s)) {
    return undefined
  }
  try {
    return JSON.parse(s)
  } catch (_) {
    return undefined
  }
}

export function isJson(value: string): boolean {
  if (value.length < 2) {
    return false
  }

  const s = value.trim()
  const res = s[0] + s[s.length - 1]
  return res === '{}'
}

export function prettyPrint(v: any): string {
  return JSON.stringify(v, null, 4)
}
