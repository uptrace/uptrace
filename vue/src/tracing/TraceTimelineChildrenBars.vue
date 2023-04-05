<template>
  <div>
    <template v-for="child in internalChildren">
      <span
        v-for="(bar, i) in child.bars"
        :key="`${child.id}-${i}`"
        :title="`${duration(bar.duration)} ${child.name}`"
        class="span-bar"
        :style="spanBarStyle(span, bar, hidden ? child.color.base : span.color.lighten)"
      ></span>
      <TraceTimelineChildrenBars
        v-if="child.children"
        :key="child.id"
        :trace="trace"
        :span="span"
        :children="child.children"
        :hidden="hidden"
      />
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { TraceSpan, UseTrace } from '@/tracing/use-trace'

// Utilities
import { duration } from '@/util/fmt/duration'
import { spanBarStyle } from '@/models/trace-span'

export default defineComponent({
  name: 'TraceTimelineChildrenBars',

  props: {
    trace: {
      type: Object as PropType<UseTrace>,
      required: true,
    },
    span: {
      type: Object as PropType<TraceSpan>,
      required: true,
    },
    children: {
      type: Array as PropType<TraceSpan[]>,
      required: true,
    },
    hidden: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const internalChildren = computed(() => {
      return props.children.filter((child) => child.kind !== 'consumer')
    })

    return { internalChildren, duration, spanBarStyle }
  },
})
</script>
