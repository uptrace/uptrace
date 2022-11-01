<template>
  <div>
    <SpansTable
      :date-range="dateRange"
      :events-mode="eventsMode"
      :loading="spans.loading"
      :spans="spans.items"
      :order="spans.order"
      :pager="spans.pager"
      class="mb-4"
      v-on="listeners"
    />

    <XPagination :pager="spans.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { useSpans } from '@/tracing/use-spans'

// Components
import SpansTable from '@/tracing/SpansTable.vue'
import { SpanChip } from '@/tracing/SpanChips.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpanListInline',
  components: { SpansTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      default: undefined,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    where: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const spans = useSpans(() => {
      const { projectId } = route.value.params
      const query = props.where + ' | ' + props.axiosParams.query
      return {
        url: `/api/v1/tracing/${projectId}/spans`,
        params: {
          ...props.axiosParams,
          query,
        },
      }
    })

    const listeners = computed(() => {
      if (!props.uql) {
        return {}
      }
      return { 'click:chip': onChipClick }
    })

    watch(
      () => props.eventsMode,
      (eventsMode) => {
        spans.order.column = eventsMode ? AttrKey.spanTime : AttrKey.spanDuration
      },
      { immediate: true },
    )

    function onChipClick(chip: SpanChip) {
      const editor = props.uql.createEditor()
      editor.where(chip.key, '=', chip.value)
      props.uql.commitEdits(editor)
    }

    return { spans, listeners, onChipClick }
  },
})
</script>

<style lang="scss" scoped></style>
