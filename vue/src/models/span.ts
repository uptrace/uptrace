import { truncate } from 'lodash'

// Utilities
import { xkey } from '@/models/otelattr'

export type AttrMap = { [key: string]: any }

export interface SpanEvent {
  name: string
  time: string
  attrs: AttrMap
}

export interface Span {
  system: string
  groupId: string

  id: string
  parentId?: string
  traceId: string

  name: string
  kind: string

  statusCode: string
  statusMessage: string

  time: string
  duration: number
  durationSelf: number

  attrs: AttrMap
  events?: SpanEvent[]
  children?: Span[]
}

export function spanName(attrs: AttrMap, maxLength = 200): string {
  if (attrs[xkey.spanSystem] === 'db:redis') {
    const stmt = attrs[xkey.dbStatement]
    if (stmt) {
      return truncate(stmt, { length: maxLength })
    }
  }
  return attrs[xkey.spanName] ?? ''
}

export function traceShowRoute(attrs: AttrMap, query: Record<string, any> = {}) {
  query.span = attrs[xkey.spanId]
  return {
    name: 'TraceShow',
    params: {
      traceId: attrs[xkey.spanTraceId],
    },
    query,
  }
}

export function spanShowRoute(span: Span) {
  return {
    name: 'SpanShow',
    params: {
      traceId: span.attrs[xkey.spanTraceId],
      spanId: span.id,
    },
  }
}

export function groupShowLink(attrs: AttrMap) {
  return {
    to: {
      name: 'GroupList',
      query: {
        time: attrs[xkey.spanTime],
        where: `${xkey.spanGroupId} = "${attrs[xkey.spanGroupId]}"`,
      },
    },
  }
}
