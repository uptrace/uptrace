<template>
  <div>
    <SpansTable
      :date-range="dateRange"
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
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { useSpans } from '@/use/spans'

// Components
import SpansTable from '@/components/SpansTable.vue'
import { SpanChip } from '@/components/SpanChips.vue'

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
        url: `/api/tracing/${projectId}/spans`,
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
