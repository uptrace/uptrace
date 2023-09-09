<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-row>
      <v-spacer />
      <v-col cols="auto">
        <DateRangePicker :date-range="dateRange" :range-days="90" />
      </v-col>
    </v-row>
    <v-row align="center">
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat color="blue lighten-5">
            <v-toolbar-title>Metrics</v-toolbar-title>

            <v-text-field
              v-model="metricStats.searchInput"
              label="Quick search: option1|option2"
              prepend-inner-icon="mdi-magnify"
              clearable
              outlined
              dense
              hide-details="auto"
              class="ml-8"
              style="max-width: 300px"
            />

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <span v-if="metricStats.hasMore">more than </span>
              <span class="font-weight-bold">{{ metricStats.items.length }}</span>
              <span> metrics</span>
            </div>
          </v-toolbar>

          <div class="pa-4">
            <v-row align="center">
              <v-col cols="auto" class="text-body-2 text--secondary"
                >Only show metrics with attributes</v-col
              >
              <v-col cols="auto">
                <v-autocomplete
                  v-model="activeAttrKeys"
                  multiple
                  :loading="attrKeysDs.loading"
                  :items="attrKeysDs.filteredItems"
                  :error-messages="attrKeysDs.errorMessages"
                  :search-input.sync="attrKeysDs.searchInput"
                  placeholder="Show all metrics"
                  solo
                  flat
                  dense
                  background-color="grey lighten-4"
                  no-filter
                  auto-select-first
                  clearable
                  hide-details="auto"
                >
                  <template #item="{ item }">
                    <v-list-item-content>
                      <v-list-item-title>
                        {{ item.text }}
                      </v-list-item-title>
                    </v-list-item-content>
                    <v-list-item-action>
                      <v-chip small>{{ item.count }}</v-chip>
                    </v-list-item-action>
                  </template>
                </v-autocomplete>
              </v-col>
            </v-row>

            <MetricsTable
              :loading="metricStats.loading"
              :metrics="metricStats.items"
              @click:item="
                activeMetric = $event
                dialog = true
              "
            />
          </div>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="dialog" max-width="1200">
      <ExploreMetric
        v-if="activeMetric"
        :date-range="dateRange"
        :metric="activeMetric"
        @click:close="dialog = false"
      />
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { useForceReload } from '@/use/force-reload'
import { UseDateRange } from '@/use/date-range'
import { useMetricStats, MetricStats } from '@/metrics/use-metrics'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MetricsTable from '@/metrics/MetricsTable.vue'
import ExploreMetric from '@/metrics/ExploreMetric.vue'

export default defineComponent({
  name: 'Explore',
  components: { DateRangePicker, MetricsTable, ExploreMetric },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore Metrics')

    const route = useRoute()
    const { forceReloadParams } = useForceReload()

    const activeAttrKeys = shallowRef<string[]>([])
    const attrKeysDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attr-keys`,
        params: {
          ...forceReloadParams.value,
        },
      }
    })

    const metricStats = useMetricStats(() => {
      return {
        ...props.dateRange.axiosParams(),
        attr_key: activeAttrKeys.value,
      }
    })

    const dialog = shallowRef(false)
    const activeMetric = shallowRef<MetricStats>()

    return {
      activeAttrKeys,
      attrKeysDs,
      metricStats,

      dialog,
      activeMetric,
    }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
