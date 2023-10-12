<template>
  <v-row align="start" dense>
    <v-col cols="auto">
      <v-btn icon title="Remove attribute" @click="$emit('click:remove')">
        <v-icon>mdi-delete</v-icon>
      </v-btn>
    </v-col>
    <v-col>
      <v-text-field v-model="attr.key" label="Attribute name" outlined dense />
    </v-col>
    <v-col>
      <v-text-field v-model="attr.value" label="Attribute value" outlined dense />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Attr } from '@/org/use-annotations'

export default defineComponent({
  name: 'AnnotationAttrRow',

  props: {
    attr: {
      type: Object as PropType<Attr>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped></style>
