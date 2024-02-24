export interface Prefix {
  prefix: string
  keys: string[]
}

export const OTHER_PREFIX = 'other'
const sep = '_'

export function buildPrefixes(keys: string[]) {
  const prefixMap = new Map<string, string[]>()
  for (let key of keys) {
    addKey(prefixMap, key)
  }

  for (let i = 1; i < 32; i++) {
    if (prefixMap.size < 10) {
      break
    }
    compactPrefixMap(prefixMap, i)
  }
  const prefixes = prefixMapToList(prefixMap)

  const otherKeys = buildOtherKeys(keys, prefixes)
  if (otherKeys.length) {
    prefixes.push({ prefix: OTHER_PREFIX, keys: otherKeys })
  }

  return prefixes
}

function addKey(prefixMap: Map<string, string[]>, key: string) {
  let prefix = key
  while (true) {
    const i = prefix.lastIndexOf(sep)
    if (i === -1) {
      return
    }

    prefix = prefix.slice(0, i)
    const keys = prefixMap.get(prefix) ?? []
    keys.push(key)
    prefixMap.set(prefix, keys)
  }
}

function compactPrefixMap(prefixMap: Map<string, string[]>, threshold = 1) {
  prefixMap.forEach((keys, prefix) => {
    if (keys.length <= 1) {
      prefixMap.delete(prefix)
      return
    }

    prefixMap.forEach((otherKeys, otherPrefix) => {
      if (otherPrefix === prefix) {
        return
      }
      if (!otherPrefix.startsWith(prefix + sep)) {
        return
      }

      if (keys.length - otherKeys.length <= threshold) {
        prefixMap.delete(otherPrefix)
      }
    })
  })
}

function prefixMapToList(prefixMap: Map<string, string[]>): Prefix[] {
  const prefixes: Prefix[] = []
  prefixMap.forEach((keys, prefix) => {
    prefixes.push({ prefix, keys })
  })
  prefixes.sort((a, b) => a.prefix.localeCompare(b.prefix))
  return prefixes
}

function buildOtherKeys(keys: string[], prefixes: Prefix[]) {
  const other = []
  for (let key of keys) {
    if (!keyMatches(key, prefixes)) {
      other.push(key)
    }
  }
  return other
}

function keyMatches(key: string, prefixes: Prefix[]) {
  for (let prefix of prefixes) {
    if (key.startsWith(prefix.prefix)) {
      return true
    }
  }
  return false
}
