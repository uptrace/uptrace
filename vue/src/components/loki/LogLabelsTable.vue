<template>
  <v-simple-table class="v-data-table--border">
    <thead class="v-data-table-header">
      <tr>
        <th>Key</th>
        <th>Value</th>
      </tr>
    </thead>

    <LogLabelsTableBody
      :labels="labels"
      show-filters
      @click:filter="$emit('click:filter', $event)"
    />

    <template v-if="Object.keys(detectedLabels).length">
      <thead class="v-data-table-header">
        <tr>
          <th colspan="2">Detected fields</th>
        </tr>
      </thead>
      <LogLabelsTableBody :labels="detectedLabels" />
    </template>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Components
import LogLabelsTableBody from '@/components/loki/LogLabelsTableBody.vue'

export default defineComponent({
  name: 'LogLabelsTable',
  components: { LogLabelsTableBody },

  props: {
    labels: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    detectedLabels: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
