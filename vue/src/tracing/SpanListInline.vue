<template>
  <div>
    <SpansTable
      :date-range="dateRange"
      :loading="spans.loading"
      :spans="spans.items"
      :is-event="isEvent"
      :order="spans.order"
      :pager="spans.pager"
      :span-list-route="spanListRoute"
      :group-list-route="groupListRoute"
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
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'SpanListInline',
  components: { SpansTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      default: undefined,
    },
    isEvent: {
      type: Boolean,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    where: {
      type: String,
      required: true,
    },
    spanListRoute: {
      type: String,
      required: true,
    },
    groupListRoute: {
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
      () => props.isEvent,
      (isEvent) => {
        spans.order.column = isEvent ? xkey.spanTime : xkey.spanDuration
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
