<template>
  <v-combobox
    :value="value"
    v-bind="$attrs"
    persistent-hint
    :loading="loading"
    :items="suggestions.filteredItems"
    :search-input.sync="suggestions.searchInput"
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
        <v-list-item-title>{{ truncate(item.text, { length: 60 }) }}</v-list-item-title>
      </v-list-item-content>
      <v-list-item-action v-if="item.hint">
        <v-list-item-action-text>{{ item.hint }}</v-list-item-action-text>
      </v-list-item-action>
    </template>
  </v-combobox>
</template>

<script lang="ts">
import { truncate } from 'lodash'
import { defineComponent, PropType } from 'vue'

// Composables
import { UseSuggestions, Suggestion } from '@/use/suggestions'

export default defineComponent({
  name: 'SimpleSuggestions',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Object as PropType<Suggestion>,
      default: undefined,
    },
    suggestions: {
      type: Object as PropType<UseSuggestions>,
      required: true,
    },
  },

  setup(props, { emit }) {
    function onInput(value: Suggestion | string | undefined) {
      if (typeof value === 'string') {
        value = { text: value }
      }
      emit('input', value)
    }

    return {
      onInput,
      truncate,
    }
  },
})
</script>

<style lang="scss" scoped></style>
