<template>
  <div class="mb-2">
    <v-row>
      <v-col
        class="py-2 grey lighten-3 cursor-pointer"
        @click="expandedInternal = !expandedInternal"
      >
        <v-icon class="mr-1">{{ expandedInternal ? 'mdi-chevron-down' : 'mdi-chevron-up' }}</v-icon>
        <span class="text--secondary text-uppercase font-weight-medium"
          >{{ category }} ({{ items.length }})</span
        >
      </v-col>
    </v-row>
    <v-row v-if="expandedInternal">
      <v-col class="pa-0 pb-4">
        <template v-for="(item, i) in items">
          <v-divider v-if="i > 0" :key="`${item.text}-divider`" />
          <SpanFacet
            :key="item.text"
            :axios-params="axiosParams"
            :value="filtersState[item.text]"
            :attr="item.text"
            :expanded="itemExpanded(item, i)"
            @input="$emit('update:filter', { attr: item.text, value: $event })"
            @click:add-query="$emit('click:add-query', $event)"
            @click:close="$emit('click:close')"
          />
        </template>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { Item, CATEGORY_CORE } from '@/tracing/query/use-span-facets'

// Components
import SpanFacet from '@/tracing/query/SpanFacet.vue'

export default defineComponent({
  name: 'SpanFacetsCategory',
  components: { SpanFacet },

  props: {
    axiosParams: {
      type: undefined as unknown as PropType<Record<string, any> | null>,
      required: true,
    },

    filtersState: {
      type: Object as PropType<Record<string, string[]>>,
      required: true,
    },
    category: {
      type: String,
      required: true,
    },
    items: {
      type: Array as PropType<Item[]>,
      required: true,
    },
    expanded: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const expandedInternal = shallowRef(props.expanded)

    function itemExpanded(item: Item, index: number): boolean {
      if (props.category === CATEGORY_CORE && index < 3) {
        return true
      }
      return item.text in props.filtersState
    }

    return { expandedInternal, itemExpanded }
  },
})
</script>

<style lang="scss" scoped></style>
