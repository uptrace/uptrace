<template>
  <PctileChart
    v-bind="$attrs"
    :annotations="annotations"
    :loading="percentiles.loading"
    :data="percentiles.data"
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
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const percentiles = usePercentiles(() => {
      const { projectId } = route.value.params
      const req = {
        url: `/api/v1/tracing/${projectId}/percentiles`,
        params: props.axiosParams,
      }
      return req
    })

    return { percentiles }
  },
})
</script>

<style lang="scss" scoped></style>
