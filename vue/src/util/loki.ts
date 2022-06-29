export const logParse = {
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
      ?.match(/[^{\}]+(?=})/g, '$1')
      ?.map((m) => m.split(','))
      ?.flat() || [], // fix removing from array
  labelAdd: (op: string, keySubtValue: string, keyValue: string): string => {
    let labelmod = op === '!=' ? keySubtValue : keyValue
    return labelmod || ''
  },
  valueAdd: (op: string, label: string, value: string, keySubtValue: string): string => {
    const [lb, val] = label.split('=')
    const values = val.split(/[""]/)[1]
    const valueAdded = `${lb}=~"${values} | ${value}"`
    let labelMod = op == '!=' ? keySubtValue : valueAdded
    return labelMod
  },
  valueRmFromLabel: (label: string, value: string): string => {
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
  valueAddToLabel: (label: string, value: string, isEquals: boolean): string => {
    const sign = isEquals ? '=' : '=~'
    const [lb, val] = label.split(sign)
    const values = val.split(/[""]/)[1]
    const labelmod = `${lb}=~"${values.trim()} | ${value.trim()}"`
    return labelmod
  },
}
export const queryParse = {
  fromLabels: (query: string, keyVal: [string, string], op: string) => {
    const [key, value] = keyVal
    const keyValue = `${key}="${value}"`
    const keySubtValue = `${key}!="${value}"`
    let queryString = ''
    let queryArr = logParse.splitLabels(query)
    queryArr?.forEach((label) => {
      const regexQuery = label.match(/([^{}=,~!]+)/gm)
      const querySplitted = logParse.splitLabels(query)
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
        const parsed = logParse.labelAdd(op, labelMod, keyValue)
        const regs = logParse.splitLabels(query).concat(parsed)
        const joined = regs.join(',')
        queryString = joined
      } // add the case for multiple exclusion regex
      else if (label.includes('=') && label.includes(key) && !label.includes(value)) {
        let labelMod = logParse.valueAddToLabel(label, value, true)
        const matches = logParse.splitLabels(query).join(',').replace(`${label}`, labelMod)
        queryString = matches
      } else if (label.includes('=~') && label.includes(key) && label.includes(value)) {
        const labelMod = logParse.valueRmFromLabel(label, value)
        const matches = logParse.splitLabels(query).join(',').replace(`${label}`, labelMod)
        queryString = matches
      } else if (
        label.includes('=~') &&
        label.includes(key.trim()) &&
        !label.includes(value.trim())
      ) {
        const labelMod = logParse.valueAddToLabel(label, value, false)
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
