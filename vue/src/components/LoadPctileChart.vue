<template>
  <PctileChart
    v-bind="$attrs"
    :loading="percentiles.loading"
    :data="percentiles.data"
    :annotations="annotations"
  />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { usePercentiles } from '@/use/percentiles'
import { Annotation } from '@/org/use-annotations'

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
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props) {
    const { route } = useRouter()

    const percentiles = usePercentiles(() => {
      const { projectId } = route.value.params
      const req = {
        url: `/internal/v1/tracing/${projectId}/percentiles`,
        params: props.axiosParams,
      }
      return req
    })

    return { percentiles }
  },
})
</script>

<style lang="scss" scoped></style>
