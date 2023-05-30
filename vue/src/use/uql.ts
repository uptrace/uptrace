import { shallowRef, reactive, computed, proxyRefs } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'

// Utilities
import { AttrKey } from '@/models/otel'
import { quote, escapeRe } from '@/util/string'

const QUERY_PART_SEP = ' | '

export interface BackendPart {
  query: string
  error?: string
  disabled?: boolean
}

export interface QueryPart {
  id: number
  query: string
  error: string
  disabled: boolean
}

export interface BackendQueryInfo {
  parts: BackendPart[]
}

export type UseUql = ReturnType<typeof useUql>

export function useUql(queryValue = '') {
  const rawMode = shallowRef(false)
  const parts = shallowRef<QueryPart[]>([])

  const query = computed({
    set(s: string) {
      parts.value = parseParts(s).map((part) => reactive(part))
    },
    get(): string {
      return formatParts(parts.value)
    },
  })
  query.value = queryValue ?? ''

  const whereQuery = computed((): string => {
    return parts.value
      .filter((part) => /where\s+/i.test(part.query))
      .map((part) => part.query)
      .join(QUERY_PART_SEP)
  })

  function addPart(part: QueryPart) {
    parts.value.push(reactive(part))
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value.slice()
  }

  function removePart(needle: QueryPart) {
    const index = parts.value.findIndex((part) => part.id === needle.id)
    parts.value.splice(index, 1)
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value.slice()
  }

  function cleanup() {
    parts.value = parts.value.filter((part) => part.query.length > 0)
  }

  function setQueryInfo(other: BackendQueryInfo) {
    // Don't remove any parts.
    parts.value.forEach((part: QueryPart, i: number) => {
      const otherPart = other.parts[i]
      if (!otherPart) {
        return
      }
      part.query = otherPart.query
      part.error = otherPart.error ?? ''
      part.disabled = otherPart.disabled ?? false
    })
  }

  function axiosParams() {
    return {
      query: query.value,
    }
  }

  function createEditor() {
    return new UqlEditor(query.value)
  }

  function commitEdits(editor: UqlEditor) {
    query.value = editor.toString()
  }

  function syncQueryParams(paramName = 'query') {
    useRouteQuery().sync({
      fromQuery(params) {
        if (paramName in params) {
          query.value = params[paramName] ?? ''
        }
      },
      toQuery() {
        return {
          [paramName]: query.value,
        }
      },
    })
  }

  return proxyRefs({
    rawMode,
    query,
    whereQuery,
    parts,

    addPart,
    removePart,
    cleanup,

    setQueryInfo,
    syncQueryParams,
    axiosParams,

    createEditor,
    commitEdits,
  })
}

export function parseParts(query: any): QueryPart[] {
  if (typeof query !== 'string' || !query) {
    return []
  }
  return split(query, QUERY_PART_SEP).map((part) => {
    return createQueryPart(part)
  })
}

export function createQueryPart(query = ''): QueryPart {
  return {
    id: Math.random() * Number.MAX_VALUE,
    query,
    error: '',
    disabled: false,
  }
}

export function formatParts(parts: QueryPart[]): string {
  return parts
    .filter((part) => part.query.length > 0)
    .map((part) => part.query)
    .join(QUERY_PART_SEP)
}

export function joinQuery(...parts: any[]): string {
  return parts
    .filter((part) => {
      if (typeof part !== 'string') {
        return false
      }
      return part.trim()
    })
    .join(QUERY_PART_SEP)
}

//------------------------------------------------------------------------------

export function createUqlEditor() {
  return new UqlEditor()
}

export class UqlEditor {
  parts: QueryPart[]

  constructor(s: any = '') {
    this.parts = parseParts(s)
  }

  toString() {
    return formatParts(this.parts)
  }

  exploreAttr(column: string, isEventSystem = false) {
    return this.add(exploreAttr(column, isEventSystem))
  }

  add(query: string) {
    for (let part of parseParts(query)) {
      const i = this.parts.findIndex((p) => p.query === part.query)
      if (i === -1) {
        this.parts.push(part)
      }
    }
    return this
  }

  where(column: string, op: string, value?: any) {
    if (value === undefined) {
      return this.replaceOrPush(
        new RegExp(`^where\\s+${escapeRe(column)}\\s+${op}$`, 'i'),
        `where ${column} ${op}`,
      )
    }
    return this.replaceOrPush(
      new RegExp(`^where\\s+${escapeRe(column)}\\s+${op}\\s+.+$`, 'i'),
      `where ${column} ${op} ${quote(value)}`,
    )
  }

  replace(re: RegExp, query: string) {
    const part = this.parts.find((part) => re.test(part.query))
    if (part) {
      part.query = query
      return true
    }
    return false
  }

  replaceOrPush(re: RegExp, query: string) {
    if (!this.replace(re, query)) {
      this.parts.push(createQueryPart(query))
    }
    return this
  }

  replaceOrUnshift(re: RegExp, query: string) {
    if (!this.replace(re, query)) {
      this.parts.unshift(createQueryPart(query))
    }
    return this
  }

  remove(s: string | RegExp) {
    let index: number

    if (typeof s === 'string') {
      index = this.parts.findIndex((part) => part.query === s)
    } else {
      index = this.parts.findIndex((part) => s.test(part.query))
    }

    if (index >= 0) {
      this.parts.splice(index, 1)
    }
  }

  addGroupBy(column: string) {
    return this.add(`group by ${column}`)
  }

  resetGroupBy(column = '') {
    this.remove(/^group by /i)
    if (column) {
      this.add(`group by ${column}`)
    }
    return this
  }
}

export function buildWhere(column: string, op: string, value?: any) {
  if (value === undefined) {
    return `where ${column} ${op}`
  }
  return `where ${column} ${op} ${quote(value)}`
}

export function exploreAttr(column: string, isEventSystem = false) {
  const ss = isEventSystem
    ? [`group by ${column}`, AttrKey.spanCountPerMin]
    : [
        `group by ${column}`,
        AttrKey.spanCountPerMin,
        AttrKey.spanErrorRate,
        `{p50,p90,p99}(${AttrKey.spanDuration})`,
      ]
  return ss.join(QUERY_PART_SEP)
}

function split(s: string, sep: string): string[] {
  return s
    .split(sep)
    .map((s) => s.trim())
    .filter((s) => s.length)
}
