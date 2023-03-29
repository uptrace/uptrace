<template>
  <div>
    <v-sheet v-for="facet in facets" :key="facet.key" outlined rounded="lg" class="mb-4">
      <v-list :value="facetedSearch.selected[facet.key]">
        <v-subheader>{{ facet.key }}</v-subheader>

        <v-list-item
          v-for="item in facet.items"
          :key="item.value"
          :value="item.value"
          dense
          @click="facetedSearch.reset(item)"
        >
          <v-list-item-action class="my-0 mr-4">
            <v-checkbox
              :input-value="facetedSearch.isSelected(item)"
              dense
              @click.stop
              @change="facetedSearch.toggle(item)"
            ></v-checkbox>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>{{ item.value }}</v-list-item-title>
          </v-list-item-content>
          <v-list-item-action class="my-0 justify-end">
            <v-list-item-action-text v-if="'count' in item" class="font-weight-bold">
              {{ item.count }}
            </v-list-item-action-text>
          </v-list-item-action>
        </v-list-item>
      </v-list>
    </v-sheet>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Facet, UseFacetedSearch } from '@/use/faceted-search'

export default defineComponent({
  name: 'SearchFacets',

  props: {
    facetedSearch: {
      type: Object as PropType<UseFacetedSearch>,
      required: true,
    },
    facets: {
      type: Array as PropType<Facet[]>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped></style>
