<template>
  <v-sheet max-height="320" :loading="loading">
    <XPlaceholder>
      <template v-if="loading" #placeholder>
        <v-skeleton-loader width="200" type="list-item@3" loading></v-skeleton-loader>
      </template>

      <v-card-text v-if="showFilter" class="mb-n2">
        <v-text-field
          v-model="search"
          placeholder="Fuzzy filter"
          outlined
          dense
          hide-details="auto"
          autofocus
          clearable
          @click.stop
        ></v-text-field>
      </v-card-text>

      <v-card-text v-if="!items.length">There are no any suggestions.</v-card-text>

      <v-list v-else dense>
        <v-list-item v-for="(item, i) in filteredItems" :key="i" @click="onClick(item)">
          <slot name="item" :item="item">
            <v-list-item-content>
              <v-list-item-title style="text-overflow: clip">
                {{ truncateMiddle(item.text, 80) }}
              </v-list-item-title>
            </v-list-item-content>
          </slot>
        </v-list-item>
      </v-list>
    </XPlaceholder>
  </v-sheet>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch } from 'vue'

// Misc
import { truncateMiddle } from '@/util/string'

interface Item {
  text: string
  value: any
}

export default defineComponent({
  name: 'SearchableList',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array,
      required: true,
    },
    numItem: {
      type: Number,
      default: undefined,
    },
    searchInput: {
      type: String,
      default: '',
    },
    returnObject: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const search = shallowRef('')

    const normalizedItems = computed((): Item[] => {
      return props.items.map((item) => {
        if (typeof item === 'string') {
          return { text: item, value: item }
        }
        return item as Item
      })
    })

    const filteredItems = computed((): Item[] => {
      let items = normalizedItems.value.slice()

      if (!search.value) {
        return items
      }

      // @ts-ignore
      return fuzzyFilter(items, search.value, { key: 'text' })
    })

    const showFilter = computed(() => {
      if (props.numItem) {
        return props.numItem > 10
      }
      return props.items.length > 10
    })

    watch(
      () => props.searchInput,
      (s) => {
        search.value = s
      },
      { immediate: true },
    )

    watch(search, (search) => {
      ctx.emit('update:searchInput', search)
    })

    function onClick(item: Item) {
      if (props.returnObject) {
        ctx.emit('input', item)
      } else {
        ctx.emit('input', item.value)
      }
    }

    return { search, filteredItems, showFilter, onClick, truncateMiddle }
  },
})
</script>

<style lang="scss" scoped></style>
