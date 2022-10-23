<template>
  <v-autocomplete
    v-autowidth="{ minWidth: '40px' }"
    :value="value"
    :items="items"
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
    <template #selection="{ index, item }">
      <div v-if="index === 2" class="v-select__selection">, {{ value.length - 2 }} more</div>
      <div v-else-if="index < 2" class="v-select__selection text-truncate">
        {{ comma(item, index) }}
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

export default defineComponent({
  name: 'StickyFilter',

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
      type: Array as PropType<string[]>,
      required: true,
    },
    paramName: {
      type: String,
      required: true,
    },
  },

  setup() {
    const menu = shallowRef(false)

    function comma(item: string, index: number): string {
      if (index > 0) {
        return ', ' + item
      }
      return item
    }

    return {
      menu,
      comma,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-select__selection {
  max-width: 100px;
}
</style>
