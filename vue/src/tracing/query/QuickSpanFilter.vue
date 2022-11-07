<template>
  <v-autocomplete
    ref="autocomplete"
    v-autowidth="{ minWidth: '40px' }"
    :value="value"
    :items="filteredItems"
    item-value="value"
    item-text="value"
    :search-input.sync="searchInput"
    no-filter
    placeholder="none"
    :prefix="`${paramName}: `"
    multiple
    clearable
    auto-select-first
    hide-details
    dense
    outlined
    background-color="light-blue lighten-5"
    class="mr-2 v-select--fit"
    @change="$emit('input', $event)"
  >
    <template #item="{ item, attrs }">
      <v-list-item
        v-bind="attrs"
        @click="
          $emit('input', [item.value])
          autocomplete.blur()
        "
      >
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="value.indexOf(item.value) >= 0"
            dense
            @click.stop="toggleValue(item.value)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>{{ item.value }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action class="my-0">
          <v-list-item-action-text><XNum :value="item.count" /></v-list-item-action-text>
        </v-list-item-action>
      </v-list-item>
    </template>
    <template #selection="{ index, item }">
      <div v-if="index === 2" class="v-select__selection">, {{ value.length - 2 }} more</div>
      <div v-else-if="index < 2" class="v-select__selection">
        {{ withComma(item, index) }}
      </div>
    </template>
    <template #no-data>
      <div>
        <v-list-item>
          <v-list-item-content>
            <v-list-item-title class="text-subtitle-1 font-weight-regular">
              To start filtering, set <code>{{ attr }}</code> attribute.
            </v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <div class="my-4 d-flex justify-center">
          <v-btn
            href="https://uptrace.dev/opentelemetry/span-naming.html#resource-attributes"
            target="_blank"
            color="primary"
          >
            <span>Open documentation</span>
            <v-icon right>mdi-open-in-new</v-icon>
          </v-btn>
        </div>
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { Item } from '@/tracing/query/use-quick-span-filters'

// Utilities
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'QuickSpanFilter',

  props: {
    value: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array as PropType<Item[]>,
      required: true,
    },
    attr: {
      type: String,
      required: true,
    },
    paramName: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const autocomplete = shallowRef()
    const searchInput = shallowRef('')

    const filteredItems = computed(() => {
      if (!searchInput.value) {
        return props.items
      }
      return fuzzyFilter(props.items, searchInput.value, { key: 'value' })
    })

    function withComma(item: Item, index: number): string {
      const value = truncateMiddle(item.value, 20)
      if (index > 0) {
        return ', ' + value
      }
      return value
    }

    function toggleValue(value: string) {
      const values = props.value.slice()
      const index = values.indexOf(value)
      if (index >= 0) {
        values.splice(index, 1)
      } else {
        values.push(value)
      }
      ctx.emit('input', values)
    }

    return {
      autocomplete,
      searchInput,
      filteredItems,
      withComma,
      toggleValue,
    }
  },
})
</script>

<style lang="scss" scoped></style>
