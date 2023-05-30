import { truncate } from 'lodash-es'

// Utilities
import { AttrKey } from '@/models/otel'

export type AttrMap = { [key: string]: any }

export interface Span {
  id: string
  parentId?: string
  traceId: string
  standalone?: boolean

  projectId: number
  groupId: string

  system: string
  kind: string

  name: string
  eventName?: string
  displayName: string

  time: string
  duration: number
  durationSelf: number

  statusCode: string
  statusMessage: string

  attrs: AttrMap
  events?: SpanEvent[]
  children?: Span[]
}

export interface SpanEvent {
  name: string
  time: string
  attrs: AttrMap

  system?: string
  groupId?: string
}

export function spanName(span: Span, maxLength = 120): string {
  if (span.system === 'db:redis') {
    const stmt = span.attrs[AttrKey.dbStatement]
    if (stmt) {
      return truncate(stmt, { length: maxLength })
    }
  }
  return truncate(span.name, { length: maxLength })
}
