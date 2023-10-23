<template>
  <v-card outlined rounded="lg" min-width="120" class="border-bottom">
    <v-row class="no-gutters">
      <v-col class="px-3 py-4">
        <v-tooltip top>
          <template #activator="{ on, attrs }">
            <div
              class="body-2 text-truncate text-uppercase blue-grey--text text--lighten-2"
              v-bind="attrs"
              v-on="on"
            >
              {{ metric.name }}
            </div>
          </template>
          <span>{{ metric.tooltip }}</span>
        </v-tooltip>

        <div class="pt-4 text-h5 text-truncate">
          <slot :metric="metric">
            <NumValue :value="metric.rate" :unit="Unit.Rate" :title="`{0} ${metric.suffix}`" />
          </slot>
          <span v-if="metric.suffix" class="ml-2 text-subtitle-1 blue-grey--text text--lighten-3">{{
            metric.suffix
          }}</span>
        </div>
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Utilities
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'SystemMetricCard',

  props: {
    metric: {
      type: Object,
      required: true,
    },
  },

  setup() {
    return { Unit }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: 6px map-get($blue, 'darken-2') solid;
}
</style>
