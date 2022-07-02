const STREAM_SELECTOR_REGEX = /{[^}]*}/

const parseLog = {
  newQuery: (keyValue: string[], op: string, tags: string[]): string => {
    const [key, value] = keyValue
    return `${tags[0] || ''}{${key}${op}"${value}"}${tags[2] || ''}`
  },
  equalLabels: (keyValue: string[], op: string, tags: string[]): string => {
    if (op === '!=') {
      return parseLog.newQuery(keyValue, op, tags)
    }

    return '{}'
  },
  formatQuery: (queryString: string, tags: string[]): string => {
    return `${tags[0] || ''}{${queryString || ''}}${tags[2] || ''}`
  },
  splitLabels: (query: string): string[] =>
    query
      ?.match(/[^{}]+(?=})/g)
      ?.map((m) => m.split(','))
      ?.flat() || [], // fix removing from array
  addLabel: (op: string, keySubtValue: string, keyValue: string): string => {
    if (op === '!=') {
      return keySubtValue
    }
    return keyValue
  },
  rmValueFromLabel: (label: string, value: string): string => {
    const [lb, val] = label?.split('=~')
    let lvalue = val?.split(/[""]/)[1]
    let values = lvalue?.split('|')
    let filtered = values?.filter((f) => f.trim() !== value?.trim())

    if (filtered?.length > 1) {
      const lvalues = filtered?.join('|')?.trim()
      return lb?.trim() + '=~' + '"' + lvalues + '"'
    }
    const lvalues = filtered?.join('')?.trim()
    return lb?.trim() + '=' + '"' + lvalues + '"'
  },
  addValueToLabel: (label: string, value: string, isEquals: boolean): string => {
    const sign = isEquals ? '=' : '=~'
    const [lb, val] = label?.split(sign)
    const values = val?.split(/[""]/)[1]
    const labelmod = `${lb}=~"${values?.trim()}|${value?.trim()}"`
    return labelmod
  },
  isEqualsQuery: (query: string, keyValue: string[]): boolean => {
    const [key, value] = keyValue
    return query === `{${key}="${value}"}`
  },
  editQuery: (query: string, keyValue: string[], op: string, tags: string[]): string => {
    if (parseLog.isEqualsQuery(query, keyValue)) {
      return parseLog.equalLabels(keyValue, op, tags)
    }

    return parseQuery.fromLabels(query, keyValue, op, tags)
  },
}
const parseQuery = {
  fromLabels: (query: string, keyVal: string[], op: string, tags: string[]): string => {
    const queryString = parseQueryLabels(keyVal, query, op)
    return parseLog.formatQuery(queryString, tags)
  },
}

function parseQueryLabels(keyVal: string[], query: string, op: string) {
  const [key, value] = keyVal
  const keyValue = `${key}="${value}"`
  const keySubtValue = `${key}!="${value}"`
  let queryArr = parseLog.splitLabels(query)
  if (!queryArr) {
    return ''
  }

  for (let label of queryArr) {
    const regexQuery = label.match(/([^{}=,~!]+)/gm)
    const querySplitted = parseLog.splitLabels(query)
    if (!regexQuery) {
      return ''
    }

    if (
      !label.includes(key?.trim()) &&
      !label.includes(value?.trim()) &&
      !querySplitted?.some((s) => s.includes(key)) &&
      !querySplitted?.some((s) => s.includes(key) && s.includes(value))
    ) {
      // add new label
      let labelMod = op === '!=' ? keySubtValue : label
      const parsed = parseLog.addLabel(op, labelMod, keyValue)
      const regs = parseLog.splitLabels(query).concat(parsed)
      return regs.join(',')
    }

    if (
      label?.includes('=') &&
      label?.split('=')?.[0]?.trim() === key?.trim() &&
      !label?.includes(value)
    ) {
      // values group from existing label
      let labelMod = parseLog.addValueToLabel(label, value, true)
      return parseLog.splitLabels(query)?.join(',')?.replace(`${label}`, labelMod)
    }

    if (
      label?.includes('=~') &&
      label?.split('=~')?.[0]?.trim() === key?.trim() &&
      label?.includes(value)
    ) {
      // filter value from existing values group from label
      const labelMod = parseLog.rmValueFromLabel(label, value)
      return parseLog.splitLabels(query).join(',').replace(`${label}`, labelMod)
    }

    if (
      label?.includes('=~') &&
      label?.split('=~')?.[0]?.trim() === key?.trim() &&
      !label?.includes(value?.trim())
    ) {
      // add value to existing values group from label
      return parseLog.addValueToLabel(label, value, false)
    }

    if (
      label?.includes('=') &&
      label?.split('=')?.[0]?.trim() === key?.trim() &&
      label?.split('"')?.[1]?.trim() === value?.trim() &&
      querySplitted?.some((s) => s === label)
    ) {
      // remove label from query
      const filtered = querySplitted?.filter((f) => f !== label)
      return filtered?.join(',')
    }
  }
  return ''
}

export function decodeQuery(query: string, key: string, value: string, op: string): string {
  const { newQuery, editQuery } = parseLog
  const isQuery = query?.match(STREAM_SELECTOR_REGEX) && query?.length > 7
  const keyValue = [key, value]
  const tags = query?.split(/[{}]/)

  if (!isQuery) {
    return newQuery(keyValue, op, tags)
  }

  return editQuery(query, keyValue, op, tags)
}
