import { truncate } from 'lodash-es'

// Utilities
import { EventName } from '@/models/otel'

export type AttrMap = { [key: string]: any }

export interface Span {
  id: string
  parentId?: string
  traceId: string

  projectId: number
  groupId: string

  system: string
  kind: string

  name: string
  eventName?: string
  standalone?: boolean

  time: string
  duration: number
  durationSelf: number

  statusCode: string
  statusMessage: string

  attrs: AttrMap
  events?: Span[]
  children?: Span[]
}

export function eventOrSpanName(span: Span, maxLength = 100): string {
  let eventName = span.eventName
  if (eventName) {
    if (eventName === EventName.Log) {
      eventName = JSON.stringify(span.attrs)
    }
    return truncate(eventName, { length: maxLength })
  }
  return spanName(span, maxLength)
}

export function spanName(span: Span, maxLength = 100): string {
  return truncate(span.name, { length: maxLength })
}
