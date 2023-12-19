<template>
  <v-combobox
    :value="value"
    v-bind="$attrs"
    persistent-hint
    :loading="dataSource.loading"
    :items="dataSource.filteredItems"
    :error-messages="dataSource.errorMessages"
    :search-input.sync="dataSource.searchInput"
    return-object
    no-filter
    auto-select-first
    clearable
    outlined
    hide-details="auto"
    class="pt-0"
    @input="onInput"
  >
    <template #item="{ item }">
      <v-list-item-content>
        <v-list-item-title>{{ truncateMiddle(item.value, 80) }}</v-list-item-title>
      </v-list-item-content>
      <v-list-item-action v-if="item.hint">
        <v-list-item-action-text>{{ item.hint }}</v-list-item-action-text>
      </v-list-item-action>
    </template>
  </v-combobox>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDataSource, Item } from '@/use/datasource'

// Misc
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'UptraceCombobox',

  props: {
    value: {
      type: Object as PropType<Item>,
      default: undefined,
    },
    dataSource: {
      type: Object as PropType<UseDataSource>,
      required: true,
    },
  },

  setup(props, { emit }) {
    function onInput(value: Item | string | undefined) {
      if (typeof value === 'string') {
        value = { value, text: value }
      }
      emit('input', value)
    }

    return {
      onInput,
      truncateMiddle,
    }
  },
})
</script>

<style lang="scss" scoped></style>
