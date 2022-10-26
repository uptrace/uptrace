import { truncate } from 'lodash-es'

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

export function eventOrSpanName(span: Span, maxLength = 200): string {
  if (span.eventName) {
    return truncate(span.eventName, { length: maxLength })
  }
  return spanName(span, maxLength)
}

export function spanName(span: Span, maxLength = 200): string {
  return truncate(span.name, { length: maxLength })
}
