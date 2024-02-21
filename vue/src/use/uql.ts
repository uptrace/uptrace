import { shallowRef, reactive, computed, watch, proxyRefs, provide, inject } from 'vue'

// Composables
import { Values } from '@/use/router'

// Misc
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

  function queryParams(paramName = 'query') {
    return {
      [paramName]: query.value,
    }
  }

  function parseQueryParams(queryParams: Values, paramName = 'query') {
    query.value = queryParams.string(paramName)
  }

  function createEditor() {
    return new QueryEditor(query.value)
  }

  function commitEdits(editor: QueryEditor) {
    query.value = editor.toString()
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
    axiosParams,

    queryParams,
    parseQueryParams,

    createEditor,
    commitEdits,
  })
}

export function parseParts(query: any): QueryPart[] {
  if (typeof query !== 'string' || !query) {
    return []
  }
  return splitQuery(query).map((part) => {
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

//------------------------------------------------------------------------------

export function createQueryEditor(query = '') {
  return new QueryEditor(query)
}

export class QueryEditor {
  parts: string[]

  constructor(query = '') {
    this.parts = splitQuery(query)
  }

  toString() {
    return joinQuery(this.parts)
  }

  exploreAttr(column: string, isSpan = false) {
    return this.add(exploreAttr(column, isSpan))
  }

  add(query: string) {
    for (let otherPart of splitQuery(query)) {
      const i = this.parts.findIndex((part) => part === otherPart)
      if (i === -1) {
        this.parts.push(otherPart)
      }
    }
    return this
  }

  where(column: string, op?: string, value?: any) {
    if (op === undefined) {
      return this.replaceOrPush(
        new RegExp(`^where\\s+${escapeRe(column)}$`, 'i'),
        `where ${column}`,
      )
    }
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

  replaceOrPush(re: RegExp, query: string) {
    if (!this.replace(re, query)) {
      this.parts.push(query)
    }
    return this
  }

  replaceOrUnshift(re: RegExp, query: string) {
    if (!this.replace(re, query)) {
      this.parts.unshift(query)
    }
    return this
  }

  replace(re: RegExp, query: string) {
    const i = this.parts.findIndex((part) => re.test(part))
    if (i >= 0) {
      this.parts[i] = query
      return true
    }
    return false
  }

  filter(fn: (part: string) => boolean) {
    this.parts = this.parts.filter(fn)
    return this
  }

  groupBy(column: string) {
    return this.add(`group by ${column}`)
  }

  resetGroupBy(column = '') {
    this.filter((part) => !/^group by /i.test(part))
    if (column) {
      this.groupBy(column)
    }
    return this
  }
}

export function splitQuery(query: string): string[] {
  return query
    .split(QUERY_PART_SEP)
    .map((part) => part.trim())
    .filter((part) => part.length)
}

export function joinQuery(parts: string[]): string {
  return parts
    .filter((part) => {
      if (typeof part !== 'string') {
        return false
      }
      return part.trim()
    })
    .join(QUERY_PART_SEP)
}

export function buildWhere(column: string, op: string, value?: any) {
  if (value === undefined) {
    return `where ${column} ${op}`
  }
  return `where ${column} ${op} ${quote(value)}`
}

export function exploreAttr(column: string, isSpan = false) {
  const ss = [`group by ${column}`, AttrKey.spanCountPerMin]
  if (isSpan) {
    ss.push(AttrKey.spanErrorRate, `{p50,p90,p99}(${AttrKey.spanDuration})`)
  } else {
    ss.push(`max(${AttrKey.spanTime})`)
  }
  return ss.join(QUERY_PART_SEP)
}

//------------------------------------------------------------------------------

const injectionKey = Symbol('query-store')

export function injectQueryStore() {
  return inject(injectionKey, undefined) ?? useQueryStore(undefined)
}

export function provideQueryStore(store: QueryStore) {
  provide(injectionKey, store)
}

export type QueryStore = ReturnType<typeof useQueryStore>

export function useQueryStore(uql: UseUql | undefined) {
  const query = shallowRef('')
  const where = shallowRef('')

  if (uql) {
    watch(
      () => uql.query,
      (queryValue) => {
        query.value = queryValue
      },
      { flush: 'sync' },
    )
    watch(
      () => uql.whereQuery,
      (whereQuery) => {
        where.value = whereQuery
      },
      { flush: 'sync' },
    )
  }

  return { query, where }
}
