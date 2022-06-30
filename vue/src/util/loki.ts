// stream selector regex
const STREAM_SELECTOR_REGEX = /{[^}]*}/

const parseLog = {
  equalLabels: (
    query: string,
    preTags: string,
    postTags: string,
    keyValue: [string, string],
    op: string,
  ) => {
    if (op === '=') {
      return `${preTags || ''}{}${postTags || ''}`
    } else if (op === '!=') {
      const [key, value] = keyValue
      return `{${key}!="${value}"}`
    } else {
      return query
    }
  },
  splitLabels: (query: string): any[] =>
    query
      ?.match(/[^{}]+(?=})/g)
      ?.map((m) => m.split(','))
      ?.flat() || [], // fix removing from array
  addLabel: (op: string, keySubtValue: string, keyValue: string): string => {
    let labelmod = op === '!=' ? keySubtValue : keyValue
    return labelmod || ''
  },
  rmValueFromLabel: (label: string, value: string): string => {
    const [lb, val] = label.split('=~')
    let lvalue = val.split(/[""]/)[1]
    let values = lvalue.split('|')
    let filtered = values.filter((f) => f.trim() !== value.trim())
    let lvalues = ''
    let opr = ''
    if (filtered.length > 1) {
      lvalues = filtered.join('|')
      opr = '=~'
    } else {
      lvalues = filtered.join('')
      opr = '='
    }
    const labelmod = lb + opr + '"' + lvalues.trim() + '"'
    return labelmod
  },
  addValueToLabel: (label: string, value: string, isEquals: boolean): string => {
    const sign = isEquals ? '=' : '=~'
    const [lb, val] = label.split(sign)
    const values = val.split(/[""]/)[1]
    const labelmod = `${lb}=~"${values.trim()} | ${value.trim()}"`
    return labelmod
  },
}
const parseQuery = {
  fromLabels: (query: string, keyVal: [string, string], op: string) => {
    const [key, value] = keyVal
    const keyValue = `${key}="${value}"`
    const keySubtValue = `${key}!="${value}"`
    let queryString = ''
    let queryArr = parseLog.splitLabels(query)
    queryArr?.forEach((label) => {
      const regexQuery = label.match(/([^{}=,~!]+)/gm)
      const querySplitted = parseLog.splitLabels(query)
      if (!regexQuery) {
        return
      }

      if (
        !label.includes(key.trim()) &&
        !label.includes(value.trim()) &&
        !querySplitted?.some((s) => s.includes(key)) &&
        !querySplitted?.some((s) => s.includes(key) && s.includes(value))
      ) {
        let labelMod = op === '!=' ? keySubtValue : label
        const parsed = parseLog.addLabel(op, labelMod, keyValue)
        const regs = parseLog.splitLabels(query).concat(parsed)
        const joined = regs.join(',')
        queryString = joined
      } // add the case for multiple exclusion regex
      else if (label.includes('=') && label.includes(key) && !label.includes(value)) {
        let labelMod = parseLog.addValueToLabel(label, value, true)
        const matches = parseLog.splitLabels(query).join(',').replace(`${label}`, labelMod)
        queryString = matches
      } else if (label.includes('=~') && label.includes(key) && label.includes(value)) {
        const labelMod = parseLog.rmValueFromLabel(label, value)
        const matches = parseLog.splitLabels(query).join(',').replace(`${label}`, labelMod)
        queryString = matches
      } else if (
        label.includes('=~') &&
        label.includes(key.trim()) &&
        !label.includes(value.trim())
      ) {
        const labelMod = parseLog.addValueToLabel(label, value, false)
        queryString = labelMod
      } else if (
        !label.includes('=~') &&
        label.includes(key) &&
        label.includes(value) &&
        querySplitted.some((s) => s === label)
      ) {
        const filtered = querySplitted.filter((f) => f !== label)
        const joined = filtered.join(',')
        queryString = joined
      }
    })
    return queryString
  },
}

export function decodeQuery(
  query: string,
  key: string,
  value: string,
  selected: boolean,
  op: string,
) {
  const { equalLabels } = parseLog
  const { fromLabels } = parseQuery
  const isQuery = query.match(STREAM_SELECTOR_REGEX) && query.length > 7
  const keyValue: [string, string] = [key, value]
  const tags = query.split(/[{}]/)
  const preTags = tags[0] || ''
  const postTags = tags[2] || ''

  if (isQuery) {
    if (query === `{${key}="${value}"}`) {
      return equalLabels(query, preTags, postTags, keyValue, op)
    } else {
      return `${preTags || ''}{${fromLabels(query, keyValue, op)}}${postTags || ''}`
    }
  } else {
    if (op === '!=') {
      return `${preTags || ''}{${key}!="${value}"}${postTags || ''}`
    } else {
      return `${preTags || ''}{${key}="${value}"}${postTags || ''}`
    }
  }
}
