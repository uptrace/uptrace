// stream selector regex
const STREAM_SELECTOR_REGEX = /{[^}]*}/

const parseLog = {
  newQuery: (keyValue: string[], op: string, tags: string[]): string => {
    const [key, value] = keyValue
    return `${tags[0] || ''}{${key}${op}"${value}"}${tags[2] || ''}`
  },
  equalLabels: (keyValue: string[], op: string, tags: string[]): string => {
    return op === '!=' ? parseLog.newQuery(keyValue, op, tags) : '{}'
  },
  splitLabels: (query: string): string[] =>
    query
      ?.match(/[^{}]+(?=})/g)
      ?.map((m) => m.split(','))
      ?.flat() || [], // fix removing from array
  addLabel: (op: string, keySubtValue: string, keyValue: string): string => {
    let labelmod = op === '!=' ? keySubtValue : keyValue
    return labelmod || ''
  },
  rmValueFromLabel: (label: string, value: string): string => {
    const [lb, val] = label?.split('=~')
    let lvalue = val?.split(/[""]/)[1]
    let values = lvalue?.split('|')
    let filtered = values?.filter((f) => f.trim() !== value?.trim())
    let lvalues = ''
    let opr = ''
    if (filtered?.length > 1) {
      lvalues = filtered?.join('|')
      opr = '=~'
    } else {
      lvalues = filtered?.join('')
      opr = '='
    }
    const labelmod = lb?.trim() + opr + '"' + lvalues?.trim() + '"'
    return labelmod
  },
  addValueToLabel: (label: string, value: string, isEquals: boolean): string => {
    const sign = isEquals ? '=' : '=~'
    const [lb, val] = label?.split(sign)
    const values = val?.split(/[""]/)[1]
    const labelmod = `${lb}=~"${values?.trim()} | ${value?.trim()}"`
    return labelmod
  },
  isEqualsQuery: (query: string, keyValue: string[]): boolean => {
    const [key, value] = keyValue
    return query === `{${key}="${value}"}`
  },
  editQuery: (query: string, keyValue: string[], op: string, tags: string[]): string => {
    return parseLog.isEqualsQuery(query, keyValue)
      ? parseLog.equalLabels(keyValue, op, tags)
      : parseQuery.fromLabels(query, keyValue, op, tags)
  },
}
const parseQuery = {
  fromLabels: (query: string, keyVal: string[], op: string, tags: string[]): string => {
    const [key, value] = keyVal
    const keyValue = `${key}="${value}"`
    const keySubtValue = `${key}!="${value}"`
    let queryString = ''
    let queryArr = parseLog.splitLabels(query)
    if (!queryArr) {
      return ''
    }
    try {
      queryArr?.forEach((label) => {
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
          let labelMod = op === '!=' ? keySubtValue : label
          const parsed = parseLog.addLabel(op, labelMod, keyValue)
          const regs = parseLog.splitLabels(query).concat(parsed)
          const joined = regs.join(',')
          queryString = joined
        } else if (
          label?.includes('=') &&
          label?.split('=')?.[0]?.trim() === key?.trim() &&
          !label?.includes(value)
        ) {
          let labelMod = parseLog.addValueToLabel(label, value, true)
          const matches = parseLog.splitLabels(query)?.join(',')?.replace(`${label}`, labelMod)
          queryString = matches
        } else if (
          label?.includes('=~') &&
          label?.split('=~')?.[0]?.trim() === key?.trim() &&
          label?.includes(value)
        ) {
          const labelMod = parseLog.rmValueFromLabel(label, value)
          const matches = parseLog.splitLabels(query).join(',').replace(`${label}`, labelMod)
          queryString = matches
        } else if (
          label?.includes('=~') &&
          label?.split('=~')?.[0]?.trim() === key?.trim() &&
          !label?.includes(value?.trim())
        ) {
          const labelMod = parseLog.addValueToLabel(label, value, false)
          queryString = labelMod
        } else if (
          label?.includes('=') &&
          label?.split('=')?.[0]?.trim() === key?.trim() &&
          label?.split('"')?.[1]?.trim() === value?.trim() &&
          querySplitted?.some((s) => s === label)
        ) {
          const filtered = querySplitted?.filter((f) => f !== label)
          const joined = filtered?.join(',')
          queryString = joined
        }
      })
      return `${tags[0] || ''}{${queryString || ''}}${tags[2] || ''}`
    } catch (e) {
      console.log(e)
      return query
    }
  },
}

export function decodeQuery(query: string, key: string, value: string, op: string): string {
  const { newQuery, editQuery } = parseLog
  const isQuery = query?.match(STREAM_SELECTOR_REGEX) && query?.length > 7
  const keyValue = [key, value]
  const tags = query?.split(/[{}]/)
  return !isQuery ? newQuery(keyValue, op, tags) : editQuery(query, keyValue, op, tags)
}
