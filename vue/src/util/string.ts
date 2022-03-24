export function escapeRe(s: string) {
  return s.replace(/[\\^$*+?.()|[\]{}]/g, '\\$&')
}

export function quote(s: any): number | string {
  if (typeof s !== 'string') {
    return JSON.stringify(s)
  }
  if (!s.length) {
    return '""'
  }

  if (s[0] === "'" && s[s.length - 1] === "'") {
    return s
  }
  if (s[0] === '"' && s[s.length - 1] === '"') {
    return s
  }

  const n = parseFloat(s)
  if (!isNaN(n) && n.toString() === s) {
    return n
  }

  return JSON.stringify(s)
}

export function truncateMiddle<T>(s: T, maxLen = 32, separator = '...'): T {
  if (typeof s !== 'string') {
    return s
  }
  if (s.length <= maxLen) {
    return s
  }

  const sepLen = separator.length,
    charsToShow = maxLen - sepLen,
    frontChars = Math.ceil(charsToShow / 2),
    backChars = Math.floor(charsToShow / 2)

  const truncated = s.substr(0, frontChars) + separator + s.substr(s.length - backChars)
  return truncated as any
}
