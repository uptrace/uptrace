<template>
  <v-sheet max-height="320" :loading="loading">
    <XPlaceholder>
      <template v-if="loading" #placeholder>
        <v-skeleton-loader width="200" type="list-item@3" loading></v-skeleton-loader>
      </template>

      <template v-else-if="!items.length" #placeholder>
        <v-card-text>There are no any suggestions.</v-card-text>
      </template>

      <v-card-text v-if="showFilter" class="mb-n2">
        <v-text-field
          v-model="search"
          placeholder="Filter"
          outlined
          dense
          hide-details="auto"
          autofocus
          clearable
        ></v-text-field>
      </v-card-text>

      <v-list dense>
        <v-list-item v-for="item in filteredItems" :key="item.value" @click="onClick(item)">
          <slot name="item" :item="item">
            <v-list-item-content>
              <v-list-item-title style="text-overflow: clip">
                {{ truncate(item.text, { length: 80 }) }}
              </v-list-item-title>
            </v-list-item-content>
          </slot>
        </v-list-item>
      </v-list>
    </XPlaceholder>
  </v-sheet>
</template>

<script lang="ts">
import { truncate } from 'lodash'
import { defineComponent, shallowRef, computed, watch } from 'vue'

interface Item {
  text: string
  value: any
}

export default defineComponent({
  name: 'XList',

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
      const items = normalizedItems.value.slice()
      if (!search.value) {
        return items
      }
      return items.filter((item) => item.text.indexOf(search.value) >= 0)
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

    return { search, filteredItems, showFilter, onClick, truncate }
  },
})
</script>

<style lang="scss" scoped></style>
