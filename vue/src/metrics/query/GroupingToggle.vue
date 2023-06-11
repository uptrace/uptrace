<template>
  <v-slide-group :value="value" multiple center-active show-arrows @change="onChange($event)">
    <v-slide-item
      v-for="(item, i) in items"
      :key="item.value"
      v-slot="{ active, toggle }"
      :value="item.value"
    >
      <v-btn
        :input-value="active"
        active-class="light-blue white--text"
        small
        depressed
        rounded
        class="text-transform-none"
        :class="{ 'ml-1': i > 0 }"
        @click="toggle"
      >
        {{ item.text }}
      </v-btn>
    </v-slide-item>
  </v-slide-group>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Item } from '@/use/datasource'

export default defineComponent({
  name: 'GroupingToggle',

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    value: {
      type: Array as PropType<string[]>,
      required: true,
    },
    items: {
      type: Array as PropType<Item[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    function onChange(value: string[]) {
      if (!props.loading) {
        ctx.emit('input', value)
      }
    }

    return { onChange }
  },
})
</script>

<style lang="scss" scoped></style>
