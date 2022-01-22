import { truncate } from 'lodash'

// Utilities
import { xkey } from '@/models/otelattr'

export type AttrMap = { [key: string]: any }

export interface Span {
  projectId: number
  system: string
  groupId: string

  traceId: string
  id: string
  parentId?: string

  name: string
  eventName?: string
  kind: string

  statusCode: string
  statusMessage: string

  time: string
  duration: number
  durationSelf: number

  attrs: AttrMap
  events?: Span[]
  children?: Span[]
}

export function eventOrSpanName(span: Span, maxLength = 200): string {
  if (span.eventName) {
    return truncate(span.eventName, { length: maxLength })
  }
  return spanName(span, maxLength)
}

export function spanName(span: Span, maxLength = 200): string {
  if (span.system === 'db:redis') {
    const stmt = span.attrs[xkey.dbStatement]
    if (stmt) {
      return truncate(stmt, { length: maxLength })
    }
  }
  return span.name
}
