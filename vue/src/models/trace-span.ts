// Misc
import { Span } from '@/models/span'
import { walkTree } from '@/models/tree'
import { ColoredSpan } from '@/models/colored-system'

export interface Trace {
  id: string
  time: string
  duration: number
}

interface LabelStyle {
  position: string
  bottom: string
  left?: string
  right?: string
}

interface SpanBar {
  startPct: number
  endPct: number
  duration: number
  durationSelf: number
}

export interface TraceSpan extends Omit<Span, 'children'>, ColoredSpan<TraceSpan> {
  level: number
  startPct: number
  endPct: number

  bars: SpanBar[]
  labelStyle: LabelStyle
}

export function traceSpans(rootSpan: Span, trace: Trace) {
  const spans = _traceSpans(rootSpan as unknown as TraceSpan, trace)
  return spans
}

function _traceSpans(root: TraceSpan, trace: Trace) {
  const spans: TraceSpan[] = []

  walkTree(root, (span: TraceSpan, parent: TraceSpan | null) => {
    span.parent = parent
    spans.push(span)

    span.level = parent ? parent.level + 1 : 0
    span.labelStyle = spanLabelStyle(span.startPct, span.endPct)
    span.bars = []

    span.bars.push({
      startPct: span.startPct,
      endPct: span.endPct,
      duration: span.duration,
      durationSelf: span.durationSelf,
    })

    return true
  })

  return spans
}

export function spanLabelStyle(startPct: number, endPct: number): LabelStyle {
  const labelStyle: LabelStyle = {
    position: 'absolute',
    bottom: '6px',
  }

  if (startPct <= 0.5) {
    labelStyle.left = pct(startPct)
  } else {
    labelStyle.right = pct(1 - endPct)
  }

  return labelStyle
}

export function spanBarStyle(span: TraceSpan, bar: SpanBar, color: string) {
  const startPct = Math.max(bar.startPct, span.startPct)
  const endPct = Math.min(bar.endPct, span.endPct)

  let width: any = endPct - startPct
  if (width >= 0.001) {
    width = pct(width)
  } else {
    width = '1px'
  }

  return { left: pct(bar.startPct), width, backgroundColor: color }
}

function pct(n: number) {
  return n * 100 + '%'
}
