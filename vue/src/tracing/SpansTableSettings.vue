<template>
  <v-dialog v-model="dialog" max-width="600">
    <template #activator="{ on, attrs }">
      <v-btn icon class="ml-2" v-bind="attrs" v-on="on">
        <v-icon>mdi-cog</v-icon>
      </v-btn>
    </template>
    <v-card>
      <v-card-title>Table settings</v-card-title>
      <v-divider />
      <v-card-text>
        <v-autocomplete
          :items="attrs"
          label="Columns"
          multiple
          auto-select-first
          chips
          deletable-chips
          outlined
          hide-details="auto"
          class="mt-4"
          @change="$emit('input', $event)"
        ></v-autocomplete>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn color="primary" text @click="dialog = false">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Utilities
import { Span } from '@/models/span'

export default defineComponent({
  name: 'SpansTableSettings',

  props: {
    spans: {
      type: Array as PropType<Span[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const dialog = shallowRef(false)

    const attrs = computed((): string[] => {
      if (!dialog.value) {
        return []
      }

      const AttrKeys: Record<string, null> = {}

      for (let span of props.spans) {
        const keys = Object.keys(span.attrs).filter((key) => !key.startsWith('_'))
        for (let key of keys) {
          AttrKeys[key] = null
        }
      }

      return Object.keys(AttrKeys).sort()
    })

    return { dialog, attrs }
  },
})
</script>

<style lang="scss" scoped></style>
