<template>
  <SparklineChart :name="column" :line="line" :time="time" />
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

// Components
import SparklineChart from '@/components/SparklineChart.vue'

export default defineComponent({
  name: 'LoadGroupSparkline',
  components: { SparklineChart },

  props: {
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    where: {
      type: String,
      required: true,
    },
    column: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const { data } = useWatchAxios(() => {
      const { projectId } = route.value.params
      const query = props.where + ' | ' + props.axiosParams.query
      return {
        url: `/api/tracing/${projectId}/stats`,
        params: {
          ...props.axiosParams,
          query,
          column: props.column,
        },
      }
    })

    const line = computed((): number[] => {
      if (data.value) {
        return data.value[props.column] ?? []
      }
      return []
    })

    const time = computed((): string[] => {
      return data.value?.time ?? []
    })

    return { line, time }
  },
})
</script>
