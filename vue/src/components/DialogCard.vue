<template>
  <v-dialog
    :value="value"
    :max-width="maxWidth"
    :fullscreen="fullscreen"
    @input="$emit('input', $event)"
  >
    <v-card flat outlined>
      <v-toolbar color="bg--none-primary" flat>
        <v-toolbar-title>{{ title }}</v-toolbar-title>
        <slot name="toolbar-append" />

        <v-spacer />

        <v-toolbar-items>
          <v-btn icon @click="$emit('input', false)">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <slot />
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'DialogCard',

  props: {
    value: {
      type: Boolean,
      required: true,
    },
    title: {
      type: String,
      required: true,
    },
    maxWidth: {
      type: [Number, String],
      default: 'auto',
    },
    fullscreen: {
      type: Boolean,
      default: false,
    },
  },
})
</script>

<style lang="scss" scoped></style>
