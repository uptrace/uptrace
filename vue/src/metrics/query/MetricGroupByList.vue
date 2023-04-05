<template>
  <SearchableList
    :loading="suggestions.loading"
    :items="suggestions.filteredItems"
    :num-item="suggestions.items.length"
    :search-input.sync="suggestions.searchInput"
    return-object
    @input="$emit('input', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { useDataSource } from '@/use/datasource'
import { Metric } from '@/metrics/types'

// Components
import SearchableList from '@/components/SearchableList.vue'

export default defineComponent({
  name: 'MetricGroupByList',
  components: { SearchableList },

  props: {
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      required: true,
    },
    metric: {
      type: Object as PropType<Metric>,
      required: true,
    },
    alias: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const suggestions = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attr-keys`,
        params: {
          ...props.axiosParams,
          metric: props.metric.name,
          alias: props.alias,
        },
      }
    })

    return { suggestions }
  },
})
</script>

<style lang="scss" scoped></style>
