<template>
  <v-list :value="facetedSearch.selected[facet.key]">
    <v-subheader>{{ facet.key }}</v-subheader>

    <v-text-field
      v-if="facet.items.length > pager.perPage"
      v-model="searchInput"
      prepend-inner-icon="mdi-magnify"
      outlined
      dense
      hide-details="auto"
      clearable
      class="mx-2 mb-2"
    ></v-text-field>

    <v-list-item
      v-for="item in pagedItems"
      :key="item.value"
      :value="item.value"
      dense
      @click="facetedSearch.toggleOne(item)"
    >
      <v-list-item-action class="my-0 mr-4">
        <v-checkbox
          :input-value="facetedSearch.isSelected(item)"
          @click.stop="facetedSearch.toggle(item)"
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

    <XPagination v-if="pager.numPage > 1" :pager="pager" total-visible="5" :show-pager="false" />
  </v-list>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { usePager } from '@/use/pager'
import { Facet, UseFacetedSearch } from '@/use/faceted-search'

export default defineComponent({
  name: 'SearchFacetList',

  props: {
    facetedSearch: {
      type: Object as PropType<UseFacetedSearch>,
      required: true,
    },
    facet: {
      type: Object as PropType<Facet>,
      required: true,
    },
  },

  setup(props) {
    const searchInput = shallowRef('')
    const pager = usePager({ perPage: 10 })

    const filteredItems = computed(() => {
      if (!searchInput.value) {
        return props.facet.items
      }

      return fuzzyFilter(props.facet.items, searchInput.value, { key: 'value' })
    })

    const pagedItems = computed(() => {
      return filteredItems.value.slice(pager.pos.start, pager.pos.end)
    })

    watch(
      () => filteredItems.value.length,
      (numItem) => {
        pager.numItem = numItem
      },
      { immediate: true },
    )

    return {
      searchInput,
      pagedItems,
      pager,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-input {
  & ::v-deep .v-input__slot {
    min-height: 32px !important;
    height: 32px !important;
  }

  & ::v-deep .v-input__prepend-inner,
  & ::v-deep .v-input__append-inner {
    margin-top: 5px !important;
  }
}
</style>
