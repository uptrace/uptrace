import { shallowRef, reactive, computed, proxyRefs, watch, ComputedRef } from '@vue/composition-api'

// Composables
import { useRouter, useQuery } from '@/use/router'

// Utilities
import { xkey } from '@/models/otelattr'
import { quote, escapeRe } from '@/util/string'

const QUERY_PART_SEP = ' | '
export const GROUP_ID_FILTER_RE = /^where\s+span\.group_id\s+=\s+(\S+)$/i

export interface QueryPart {
  query: string
  error?: string
  disabled?: boolean
}

export interface Part {
  query: string
  error: string
  disabled: boolean
}

interface UqlConfig {
  query?: string | ComputedRef<string>
  paramName?: string
  syncQuery?: boolean
}

export type UseUql = ReturnType<typeof useUql>

export function useUql(cfg: UqlConfig = {}) {
  const paramName = cfg.paramName ?? 'query'

  const { router, route } = useRouter()
  const rawMode = shallowRef(false)
  const parts = shallowRef<Part[]>([])

  const query = computed({
    set(s: string) {
      parts.value = parseParts(s).map((part) => reactive(part))
    },
    get(): string {
      return formatParts(parts.value)
    },
  })

  function addPart(part: Part) {
    parts.value.push(reactive(part))
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value
  }

  function removePart(index: number) {
    parts.value.splice(index, 1)
    // eslint-disable-next-line no-self-assign
    parts.value = parts.value
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
    if (!cfg.syncQuery) {
      query.value = editor.toString()
      return
    }

    router
      .push({
        query: {
          ...route.value.query,
          [paramName]: editor.toString(),
        },
      })
      .catch(() => {})
  }

  if (typeof cfg.query === 'string') {
    query.value = cfg.query
  } else if (cfg.query) {
    watch(
      cfg.query as ComputedRef<string>,
      (queryValue) => {
        query.value = queryValue
      },
      { immediate: true },
    )
  }

  if (cfg.syncQuery) {
    useQuery().sync({
      fromQuery(params) {
        let s = params[paramName] ?? cfg.query ?? ''
        if (params.where) {
          s += QUERY_PART_SEP + 'where ' + params.where
        }
        query.value = s
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

export function parseParts(query: any): Part[] {
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

export function createPart(query = ''): Part {
  return {
    query,
    error: '',
    disabled: false,
  }
}

export function formatParts(parts: Part[]): string {
  return parts
    .filter((part) => part.query.length > 0)
    .map((part) => part.query)
    .join(QUERY_PART_SEP)
}

export class UqlEditor {
  parts: Part[]

  constructor(s: any = '') {
    this.parts = parseParts(s)
  }

  toString() {
    return formatParts(this.parts)
  }

  reset() {
    this.parts = []
    this.add(buildGroupBy(xkey.spanGroupId))
    return this
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

  replaceGroupBy(column: string) {
    return this.replaceOrUnshift(/^group\s+by\s+/i, `group by ${column}`)
  }
}

export function buildWhere(column: string, op: string, value?: any) {
  if (value === undefined) {
    return `${column} ${op}`
  }
  return `${column} ${op} ${quote(value)}`
}

export function buildGroupBy(column: string) {
  const ss = [
    `group by ${column}`,
    xkey.spanCountPerMin,
    xkey.spanErrorPct,
    `{p50,p90,p99}(${xkey.spanDuration})`,
  ]
  return ss.join(QUERY_PART_SEP)
}
