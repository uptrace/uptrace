import { orderBy } from 'lodash-es'
import { ref, computed, watch, proxyRefs, shallowReactive, set } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'

import { injectForceReload } from '@/use/force-reload'
import { traceSpans, Trace, TraceSpan } from '@/models/trace-span'
import { SpanEvent } from '@/models/span'
import { spanColoredSystems, ColoredSystem } from '@/models/colored-system'

// Utilities
import { walkTree } from '@/models/tree'

export type { TraceSpan }

export type UseTrace = ReturnType<typeof useTrace>

export function useTrace(axiosParams: AxiosParamsSource) {
  const { route } = useRouter()
  const forceReload = injectForceReload()

  const activeSpanId = ref('')
  const coloredSystems = ref<ColoredSystem[]>([])
  const { isVisible, isExpanded, toggleTree, showTree } = useHiddenSpans()

  const { status, loading, error, data } = useWatchAxios(() => {
    const { projectId, traceId } = route.value.params
    return {
      url: `/internal/v1/tracing/${projectId}/traces/${traceId}`,
      params: {
        ...forceReload.params,
        ...axiosParams(),
      },
    }
  })

  const id = computed((): string | undefined => {
    return route.value.params.traceId
  })

  const trace = computed((): Trace | undefined => {
    return data.value?.trace
  })

  const root = computed((): TraceSpan | undefined => {
    return data.value?.root
  })

  const hasMore = computed((): boolean => {
    return data.value?.hasMore ?? false
  })

  const spans = computed((): TraceSpan[] => {
    const root = data.value?.root
    if (!root) {
      return []
    }

    const spans = traceSpans(root, trace.value!)
    return spans
  })

  const activeSpan = computed((): TraceSpan | undefined => {
    if (!activeSpanId.value) {
      return
    }
    return spans.value.find((span) => span.id === activeSpanId.value)
  })

  const events = computed((): Record<string, SpanEvent[]> => {
    const eventMap: Record<string, SpanEvent[]> = {}

    for (let span of spans.value) {
      if (!span.events) {
        continue
      }

      for (let event of span.events) {
        if (!event.system) {
          continue
        }
        let arr = eventMap[event.system]
        if (!arr) {
          arr = []
          eventMap[event.system] = arr
        }

        arr.push(event)
      }
    }

    for (let system in eventMap) {
      eventMap[system] = orderBy(eventMap[system], (event) => event.time)
    }

    return eventMap
  })

  watch(
    spans,
    (spans) => {
      if (!spans.length) {
        return
      }

      const root = spans[0]
      if (root.children) {
        for (let child of root.children) {
          walkTree(child, (span) => {
            toggleTree(span)
            return true
          })
        }
      }
    },
    { immediate: true },
  )

  watch(activeSpan, (span) => {
    if (span) {
      showTree(span)
    }
  })

  watch(
    root,
    (root) => {
      if (root) {
        coloredSystems.value = spanColoredSystems(root)
      }
    },
    { flush: 'sync' },
  )

  return proxyRefs({
    activeSpanId,
    activeSpan,

    status,
    loading,
    error,

    coloredSystems,
    root,
    hasMore,

    spans,
    id,
    trace,
    events,

    isVisible,
    isExpanded,
    toggleTree,
    showTree,
  })
}

function useHiddenSpans() {
  const hidden = shallowReactive<Record<string, number | undefined>>({})
  const collapsed = shallowReactive<Record<string, boolean | undefined>>({})

  function isVisible(span: TraceSpan): boolean {
    if (!(span.id in hidden)) {
      set(hidden, span.id, 0)
    }
    return !hidden[span.id]
  }

  function isExpanded(span: TraceSpan): boolean {
    if (!(span.id in collapsed)) {
      set(collapsed, span.id, false)
    }
    return !collapsed[span.id]
  }

  function toggleTree(span: TraceSpan): void {
    set(collapsed, span.id, isExpanded(span))
    hideChildren(span, isExpanded(span) ? 1 : -1)
  }

  function hideChildren(span: TraceSpan, n = 1) {
    if (!span.children) {
      return
    }
    for (let s of span.children) {
      set(hidden, s.id, (hidden[s.id] ?? 0) + n)
      hideChildren(s, n)
    }
  }

  function showTree(span: TraceSpan | null): void {
    span = span!.parent
    while (span) {
      if (!isExpanded(span)) {
        toggleTree(span)
      }
      span = span.parent
    }
  }

  return { isVisible, isExpanded, toggleTree, showTree }
}
