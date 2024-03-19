<template>
  <div>
    <v-row>
      <v-col
        class="py-2 cursor-pointer"
        :class="$vuetify.theme.isDark ? 'grey darken-4' : 'grey lighten-3'"
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
          <v-divider v-if="i > 0" :key="`${item.value}-divider`" />
          <FacetItem
            :key="item.value"
            :component="component"
            :axios-params="axiosParams"
            :value="filtersState[item.value]"
            :attr="item.value"
            :pending="pinnedFacetMan.pending"
            :expanded="itemExpanded(item, i)"
            :pinned="item.pinned"
            @input="$emit('update:filter', { attr: item.value, op: '=', value: $event })"
            @update:filter="$emit('update:filter', $event)"
            @click:close="$emit('click:close')"
            @click:pin="pinFacet(item.value)"
            @click:unpin="unpinFacet(item.value)"
          />
        </template>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { Item, Category } from '@/components/facet/types'
import { usePinnedFacetManager } from '@/components/facet/use-pinned-facets'

// Components
import FacetItem from '@/components/facet/FacetItem.vue'

export default defineComponent({
  name: 'FacetItemsCategory',
  components: { FacetItem },

  props: {
    component: {
      type: String,
      required: true,
    },
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

  setup(props, ctx) {
    const pinnedFacetMan = usePinnedFacetManager()
    const expandedInternal = shallowRef(props.expanded)

    function pinFacet(attr: string) {
      pinnedFacetMan.add(attr).then(() => {
        ctx.emit('update:pinned')
      })
    }

    function unpinFacet(attr: string) {
      pinnedFacetMan.remove(attr).then(() => {
        ctx.emit('update:pinned')
      })
    }

    function itemExpanded(item: Item, index: number): boolean {
      switch (props.category) {
        case Category.All:
          return index < 5
        case Category.Pinned:
          return index < 3
        default:
          return item.value in props.filtersState
      }
    }

    return {
      expandedInternal,
      itemExpanded,

      pinnedFacetMan,
      pinFacet,
      unpinFacet,
    }
  },
})
</script>

<style lang="scss" scoped></style>
