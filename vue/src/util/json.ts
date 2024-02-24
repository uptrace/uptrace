export function parseJson(value: any): any {
  if (!value) {
    return false
  }
  if (typeof value === 'object' && !Array.isArray(value)) {
    return value
  }
  if (typeof value !== 'string') {
    return undefined
  }

  if (!isJson(value)) {
    return undefined
  }

  try {
    return JSON.parse(value)
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
  return JSON.stringify(v, null, 2)
}
