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
