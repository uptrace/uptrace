<template>
  <div>
    <v-sheet v-for="facet in facets" :key="facet.key" outlined rounded="lg" class="mb-4">
      <SearchFacetList :faceted-search="facetedSearch" :facet="facet" />
    </v-sheet>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Facet, UseFacetedSearch } from '@/use/faceted-search'

// Components
import SearchFacetList from '@/components/SearchFacetList.vue'

export default defineComponent({
  name: 'SearchFacets',
  components: { SearchFacetList },

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
