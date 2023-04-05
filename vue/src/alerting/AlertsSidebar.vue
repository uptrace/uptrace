<template>
  <div>
    <v-text-field
      v-model="facetedSearch.q"
      outlined
      dense
      prepend-inner-icon="mdi-magnify"
      placeholder="Search alerts..."
      clearable
    />

    <SearchFacets :facets="facets" :faceted-search="facetedSearch" />

    <div v-if="facetedSearch.selectedLength" class="text-center">
      <v-btn text @click="facetedSearch.resetAll">Reset all</v-btn>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseFacetedSearch, Facet } from '@/use/faceted-search'

// Components
import SearchFacets from '@/components/SearchFacets.vue'

export default defineComponent({
  name: 'AlertsSidebar',
  components: { SearchFacets },

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

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped>
.v-btn-toggle ::v-deep .v-btn {
  padding: 0 10px !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
}
</style>
