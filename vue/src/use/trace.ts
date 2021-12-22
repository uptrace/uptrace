import { ref, computed, watch, proxyRefs, shallowReactive, set } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'
import { useForceReload } from '@/use/force-reload'
import { traceSpans, Trace, TraceSpan } from '@/models/trace-span'
import { spanColoredSystems, ColoredSystem } from '@/models/colored-system'

// Utilities
import { walkTree } from '@/models/tree'

export type { TraceSpan }

export type UseTrace = ReturnType<typeof useTrace>

export function useTrace() {
  const { route } = useRouter()
  const { forceReloadParams } = useForceReload()

  const activeSpanId = ref('')
  const coloredSystems = ref<ColoredSystem[]>([])
  const { isVisible, isExpanded, toggleTree, showTree } = useHiddenSpans()

  const { loading, data, error } = useWatchAxios(() => {
    const { traceId } = route.value.params
    return {
      url: `/api/tracing/traces/${traceId}`,
      params: {
        ...forceReloadParams.value,
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

    loading,
    error,

    coloredSystems,
    root,
    spans,
    id,
    trace,

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
