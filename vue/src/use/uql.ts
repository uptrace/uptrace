import { shallowRef, reactive, computed, proxyRefs } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'

// Utilities
import { AttrKey } from '@/models/otel'
import { quote, escapeRe } from '@/util/string'

const QUERY_PART_SEP = ' | '
export const GROUP_ID_FILTER_RE = /^where\s+span\.group_id\s+=\s+(\S+)$/i

export interface QueryPart {
  query: string
  error: string
  disabled: boolean
}

interface UqlConfig {
  query?: string
  paramName?: string
  syncQuery?: boolean
}

export type UseUql = ReturnType<typeof useUql>

export function useUql(cfg: UqlConfig = {}) {
  const paramName = cfg.paramName ?? 'query'

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

  const whereQuery = computed((): string => {
    return parts.value
      .filter((part) => /where\s+/i.test(part.query))
      .map((part) => part.query)
      .join(QUERY_PART_SEP)
  })

  query.value = cfg.query ?? ''

  if (cfg.syncQuery) {
    useRouteQuery().sync({
      fromQuery(params) {
        if (params[paramName]) {
          query.value = params[paramName]
        }
      },
      toQuery() {
        return {
          [paramName]: query.value,
        }
      },
    })
  }

  function addPart(part: QueryPart) {
    parts.value.push(reactive(part))
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value.slice()
  }

  function removePart(index: number) {
    parts.value.splice(index, 1)
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value.slice()
  }

  function cleanup() {
    parts.value = parts.value.filter((part) => part.query.length > 0)
  }

  function syncParts(other: QueryPart[]) {
    parts.value.forEach((part: QueryPart, i: number) => {
      const otherPart = other[i]
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

  return proxyRefs({
    rawMode,
    query,
    whereQuery,
    parts,

    addPart,
    removePart,
    cleanup,

    syncParts,
    axiosParams,

    createEditor,
    commitEdits,
  })
}

export function parseParts(query: any): QueryPart[] {
  if (typeof query !== 'string' || !query) {
    return []
  }

  return query
    .split(QUERY_PART_SEP)
    .map((s) => s.trim())
    .filter((s) => s.length)
    .map((s) => {
      return createPart(s)
    })
}

export function createPart(query = ''): QueryPart {
  return {
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
      this.parts.push(createPart(query))
    }
    return this
  }

  replaceOrUnshift(re: RegExp, query: string) {
    if (!this.replace(re, query)) {
      this.parts.unshift(createPart(query))
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

  groupId() {
    for (let part of this.parts) {
      const m = part.query.match(GROUP_ID_FILTER_RE)
      if (m) {
        return m[1]
      }
    }
    return ''
  }

  removeGroupFilter() {
    this.remove(GROUP_ID_FILTER_RE)
    return this
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
        AttrKey.spanErrorPct,
        `{p50,p90,p99}(${AttrKey.spanDuration})`,
      ]
  return ss.join(QUERY_PART_SEP)
}
