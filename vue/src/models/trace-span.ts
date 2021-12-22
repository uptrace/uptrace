// Utilities
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

interface BarStyle {
  left: string
  width: string
}

interface ColoredBarStyle extends BarStyle {
  backgroundColor?: string
}

interface SpanBar {
  duration: number
  style: BarStyle
  coloredStyle: ColoredBarStyle
  lightenStyle: ColoredBarStyle
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
    const spanStartPct = span.startPct
    const durationPct = spanDurationPct(spanStartPct, span.duration, trace.duration)
    span.endPct = spanStartPct + durationPct

    span.labelStyle = spanLabelStyle(spanStartPct, durationPct)
    span.bars = []

    const style = spanBarStyle(spanStartPct, durationPct)
    span.bars.push({
      duration: span.durationSelf,
      style,
      coloredStyle: Object.assign({}, style, { backgroundColor: span.color }),
      lightenStyle: Object.assign({}, style, { backgroundColor: span.lightenColor }),
    })

    return true
  })

  return spans
}

export function spanDurationPct(startPct: number, duration: number, traceDuration: number): number {
  let durationPct = duration / traceDuration
  if (startPct + durationPct > 1) {
    return 1 - startPct
  }
  return durationPct
}

export function spanBarStyle(startPct: number, durationPct: number): BarStyle {
  let left = pct(startPct)
  let width = durationPct <= 0.001 ? '1px' : pct(durationPct)
  return { left, width }
}

export function spanLabelStyle(startPct: number, durationPct: number): LabelStyle {
  const labelStyle: LabelStyle = {
    position: 'absolute',
    bottom: '6px',
  }

  if (startPct <= 0.5) {
    labelStyle.left = pct(startPct)
  } else {
    let right = 1 - (startPct + durationPct)
    labelStyle.right = pct(right)
  }

  return labelStyle
}

function pct(n: number) {
  return n * 100 + '%'
}
