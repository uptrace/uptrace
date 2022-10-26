<template>
  <SearchableList
    :loading="suggestions.loading"
    :items="suggestions.filteredItems"
    :num-item="suggestions.items.length"
    :search-input.sync="suggestions.searchInput"
    return-object
    @input="$emit('click:where', $event)"
  >
    <template #item="{ item }">
      <v-list-item-content>
        <v-list-item-title>
          {{ truncateMiddle(item.text, 80) }}
        </v-list-item-title>
      </v-list-item-content>

      <v-list-item-action class="my-0" @click.stop="$emit('click:where-not', item)">
        <v-btn icon>
          <v-icon small>mdi-not-equal</v-icon>
        </v-btn>
      </v-list-item-action>
    </template>
  </SearchableList>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { useSuggestions } from '@/use/suggestions'
import { Metric } from '@/metrics/types'

// Components
import SearchableList from '@/components/SearchableList.vue'

// Utilities
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'WhereSuggestions',
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

    const suggestions = useSuggestions(() => {
      const { projectId } = route.value.params
      const metricId = props.metric.id
      return {
        url: `/api/v1/metrics/${projectId}/${metricId}/where`,
        params: {
          ...props.axiosParams,
          alias: props.alias,
        },
      }
    })

    return {
      suggestions,

      truncateMiddle,
    }
  },
})
</script>

<style lang="scss" scoped></style>
