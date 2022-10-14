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
  </v-autocomplete>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

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

  setup(props) {
    const menu = shallowRef(false)

    const attrs = computed(() => {
      if (props.outlined) {
        return { outlined: true }
      }
      return { dark: true, class: 'blue darken-1 elevation-5' }
    })

    return {
      menu,
      attrs,
    }
  },
})
</script>

<style lang="scss" scoped></style>
