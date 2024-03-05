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

            <v-col cols="auto">
              <v-text-field
                v-model="metrics.searchInput"
                placeholder="Quick search: option1|option2"
                prepend-inner-icon="mdi-magnify"
                clearable
                outlined
                dense
                hide-details="auto"
                style="min-width: 300px"
              />
            </v-col>
            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <span v-if="metrics.hasMore">more than </span>
              <span class="font-weight-bold">{{ metrics.items.length }}</span>
              <span> metrics</span>
            </div>
          </v-toolbar>

          <div class="pa-4">
            <v-row class="mb-4">
              <v-col>
                <v-autocomplete
                  v-model="activeAttrKeys"
                  multiple
                  :loading="attrKeysDs.loading"
                  :items="attrKeysDs.filteredItems"
                  :error-messages="attrKeysDs.errorMessages"
                  :search-input.sync="attrKeysDs.searchInput"
                  placeholder="Filter by attributes presence"
                  outlined
                  dense
                  no-filter
                  auto-select-first
                  clearable
                  hide-details="auto"
                  style="min-width: 300px"
                >
                  <template #item="{ item }">
                    <v-list-item-action class="my-0 mr-4">
                      <v-checkbox :input-value="activeAttrKeys.includes(item.value)"></v-checkbox>
                    </v-list-item-action>
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
              <v-col>
                <v-autocomplete
                  v-model="activeInstruments"
                  multiple
                  :loading="instrumentDs.loading"
                  :items="instrumentDs.filteredItems"
                  :error-messages="instrumentDs.errorMessages"
                  :search-input.sync="instrumentDs.searchInput"
                  placeholder="Filter by instruments"
                  outlined
                  dense
                  no-filter
                  auto-select-first
                  clearable
                  hide-details="auto"
                  style="min-width: 300px"
                >
                  <template #item="{ item }">
                    <v-list-item-action class="my-0 mr-4">
                      <v-checkbox
                        :input-value="activeInstruments.includes(item.value)"
                      ></v-checkbox>
                    </v-list-item-action>
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
              <v-col>
                <v-autocomplete
                  v-model="activeOtelLibraries"
                  multiple
                  :loading="otelLibraryDs.loading"
                  :items="otelLibraryDs.filteredItems"
                  :error-messages="otelLibraryDs.errorMessages"
                  :search-input.sync="otelLibraryDs.searchInput"
                  placeholder="Filter by instrumentation libraries"
                  outlined
                  dense
                  no-filter
                  auto-select-first
                  clearable
                  hide-details="auto"
                  style="min-width: 300px"
                >
                  <template #item="{ item }">
                    <v-list-item-action class="my-0 mr-4">
                      <v-checkbox
                        :input-value="activeOtelLibraries.includes(item.value)"
                      ></v-checkbox>
                    </v-list-item-action>
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
              :loading="metrics.loading"
              :metrics="metrics.items"
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
        v-if="dialog && activeMetric"
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
import { UseDateRange } from '@/use/date-range'
import { useExploreMetrics, ExploredMetric } from '@/metrics/use-metrics'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MetricsTable from '@/metrics/MetricsTable.vue'
import ExploreMetric from '@/metrics/ExploreMetric.vue'

// Misc
import { AttrKey } from '@/models/otel'

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

    const activeAttrKeys = shallowRef<string[]>([])
    const activeInstruments = shallowRef<string[]>([])
    const activeOtelLibraries = shallowRef<string[]>([])

    const attrKeysDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attributes`,
        params: {
          ...props.dateRange.axiosParams(),
          instrument: activeInstruments.value,
          otel_library_name: activeOtelLibraries.value,
        },
      }
    })

    const instrumentDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attributes/${AttrKey.metricInstrument}`,
        params: {
          ...props.dateRange.axiosParams(),
          attr_key: activeAttrKeys.value,
          otel_library_name: activeOtelLibraries.value,
        },
      }
    })

    const otelLibraryDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attributes/${AttrKey.otelLibraryName}`,
        params: {
          ...props.dateRange.axiosParams(),
          attr_key: activeAttrKeys.value,
          instrument: activeInstruments.value,
        },
      }
    })

    const metrics = useExploreMetrics(() => {
      return {
        ...props.dateRange.axiosParams(),
        attr_key: activeAttrKeys.value,
        instrument: activeInstruments.value,
        otel_library_name: activeOtelLibraries.value,
      }
    })

    const dialog = shallowRef(false)
    const activeMetric = shallowRef<ExploredMetric>()

    return {
      activeAttrKeys,
      attrKeysDs,
      activeInstruments,
      instrumentDs,
      activeOtelLibraries,
      otelLibraryDs,
      metrics,

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
