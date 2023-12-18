<template>
  <div>
    <v-row align="center" dense class="mb-3">
      <v-col cols="3" md="2">
        <v-text-field
          v-model="searchInput"
          placeholder="Filter metrics"
          clearable
          dense
          outlined
          hide-details
          class="pt-0"
        />
      </v-col>

      <v-col cols="9" md="10">
        <v-slide-group v-model="activePrefix" center-active show-arrows class="ml-2">
          <v-slide-item
            v-for="(item, i) in prefixes"
            :key="item.prefix"
            v-slot="{ active, toggle }"
            :value="item"
          >
            <v-btn
              :input-value="active"
              active-class="light-blue white--text"
              small
              depressed
              rounded
              :class="{ 'ml-1': i > 0 }"
              @click="toggle"
            >
              {{ item.prefix }}
            </v-btn>
          </v-slide-item>
        </v-slide-group>
      </v-col>
    </v-row>

    <v-row dense>
      <v-col v-for="(metric, i) in filteredMetrics" :key="metric.name" cols="12" md="6">
        <v-expansion-panels :value="i < 20 ? 0 : undefined">
          <v-expansion-panel>
            <v-expansion-panel-header :title="metric.description" class="user-select-text">
              <div>
                <span class="text-subtitle-2">{{ metric.name }} ({{ metric.numTimeseries }})</span>
                <v-chip
                  v-if="metric.instrument"
                  label
                  color="grey lighten-4"
                  title="Instrument"
                  class="ml-2"
                  >{{ metric.instrument }}</v-chip
                >
                <v-chip v-if="metric.unit" label color="grey lighten-4" title="Unit" class="ml-2">{{
                  metric.unit
                }}</v-chip>
              </div>
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <GroupMetricsItem :date-range="dateRange" :metric="metric" :where="where" />
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import GroupMetricsItem from '@/metrics/GroupMetricsItem.vue'

// Misc
import { Metric } from '@/metrics/types'
import { buildPrefixes, Prefix } from '@/models/key-prefixes'

export default defineComponent({
  name: 'GroupMetrics',
  components: { GroupMetricsItem },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    where: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const searchInput = shallowRef('')
    const activePrefix = shallowRef<Prefix>()
    const prefixes = computed(() => {
      const keys = props.metrics.map((metric) => metric.name)
      return buildPrefixes(keys)
    })

    const filteredMetrics = computed((): Metric[] => {
      let metrics = props.metrics

      if (activePrefix.value) {
        metrics = props.metrics.filter((metric) => {
          return activePrefix.value!.keys.includes(metric.name)
        })
      }

      if (searchInput.value) {
        metrics = fuzzyFilter(metrics, searchInput.value, { key: 'name' })
      }

      return metrics
    })

    return {
      searchInput,
      activePrefix,
      prefixes,
      filteredMetrics,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-expansion-panel-content ::v-deep .v-expansion-panel-content__wrap {
  padding-left: 2px !important;
  padding-right: 2px !important;
}
</style>
