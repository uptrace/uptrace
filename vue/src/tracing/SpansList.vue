<template>
  <div>
    <SpansTable
      :loading="spans.loading"
      :spans="spans.items"
      :order="spans.order"
      :pager="spans.pager"
      :events-mode="eventsMode"
      class="mb-4"
      v-on="listeners"
    />

    <XPagination :pager="spans.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useSpans } from '@/tracing/use-spans'
import { UseUql } from '@/use/uql'

// Components
import SpansTable from '@/tracing/SpansTable.vue'
import { SpanChip } from '@/tracing/SpanChips.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpansList',
  components: { SpansTable },

  props: {
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
    system: {
      type: String,
      default: undefined,
    },
    where: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const { route } = useRouter()

    const spans = useSpans(
      () => {
        const query = [props.axiosParams.query, props.where].filter((s) => s).join(' | ')
        const params: Record<string, any> = {
          ...props.axiosParams,
          query,
        }

        if (props.system) {
          params.system = props.system
        }

        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/spans`,
          params,
        }
      },
      {
        order: {
          column: props.eventsMode ? AttrKey.spanTime : AttrKey.spanDuration,
          desc: true,
        },
      },
    )

    const listeners = computed(() => {
      if (!props.uql) {
        return {}
      }
      return { 'click:chip': onChipClick }
    })

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
