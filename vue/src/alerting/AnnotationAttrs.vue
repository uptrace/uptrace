<template>
  <div>
    <template v-for="(value, key) in attrs">
      <v-tooltip :key="key" top>
        <template #activator="{ on, attrs }">
          <v-chip :small="small" label class="mr-1" v-bind="attrs" v-on="on">
            {{ value ? truncate(value, { length: 20 }) : key }}
          </v-chip>
        </template>
        {{ key }}: {{ value }}
      </v-tooltip>
    </template>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, PropType } from 'vue'

export default defineComponent({
  name: 'AnnotationAttrs',

  props: {
    attrs: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    small: {
      type: Boolean,
      default: false,
    },
  },

  setup() {
    return {
      truncate,
    }
  },
})
</script>

<style lang="scss" scoped></style>
