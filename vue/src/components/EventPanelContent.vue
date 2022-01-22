<template>
  <div>
    <PctileChart v-if="false" :loading="percentiles.loading" :data="percentiles.data" />
    <AttrTable :date-range="dateRange" :attrs="event.attrs" class="mt-4" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { usePercentiles } from '@/use/percentiles'

// Components
import PctileChart from '@/components/PctileChart.vue'
import AttrTable from '@/components/AttrTable.vue'

// Utilities
import { Span } from '@/models/span'

export default defineComponent({
  name: 'EventPanelContent',
  components: { PctileChart, AttrTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    event: {
      type: Object as PropType<Span>,
      required: true,
    },
  },

  setup(props) {
    const percentiles = usePercentiles(() => {
      return {
        url: `/api/v1/tracing/percentiles`,
        params: {
          group_id: undefined,
          ...props.dateRange.axiosParams(),
        },
      }
    })

    return { percentiles }
  },
})
</script>

<style lang="scss" scoped></style>
