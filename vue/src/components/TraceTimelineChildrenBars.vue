<template>
  <div>
    <template v-for="child in internalChildren">
      <span
        v-for="(bar, i) in child.bars"
        :key="`${child.id}-${i}`"
        :title="`${durationFixed(bar.duration)} ${child.name}`"
        class="span-bar"
        :style="hidden ? bar.coloredStyle : bar.lightenStyle"
      ></span>
      <TraceTimelineChildrenBars
        :key="child.id"
        v-if="child.children"
        :trace="trace"
        :span="span"
        :children="child.children"
        :hidden="hidden"
      />
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { TraceSpan, UseTrace } from '@/use/trace'

// Utilities
import { durationFixed } from '@/util/fmt/duration'

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
      return props.children.filter((child) => {
        return child.endPct <= props.span.endPct && child.kind !== 'consumer'
      })
    })
    return { internalChildren, durationFixed }
  },
})
</script>
