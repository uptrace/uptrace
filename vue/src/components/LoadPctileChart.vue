<template>
  <PctileChart v-bind="$attrs" :loading="percentiles.loading" :data="percentiles.data" />
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Composables
import { usePercentiles } from '@/use/percentiles'

// Components
import PctileChart from '@/components/PctileChart.vue'

export default defineComponent({
  name: 'LoadPctileChart',
  components: { PctileChart },

  props: {
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const percentiles = usePercentiles(() => {
      return {
        url: `/api/tracing/percentiles`,
        params: props.axiosParams,
      }
    })

    return { percentiles }
  },
})
</script>

<style lang="scss" scoped></style>
