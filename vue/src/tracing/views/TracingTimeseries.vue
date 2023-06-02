<template>
  <div>
    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat dense color="light-blue lighten-5">
            <v-toolbar-title>
              <span>Timeseries</span>
            </v-toolbar-title>

            <v-spacer />
          </v-toolbar>

          <v-container fluid class="py-4">
            <template v-if="!timeseries.status.isResolved()">
              <v-row v-for="i in 2" :key="i">
                <v-col cols="6">
                  <v-skeleton-loader type="image" />
                </v-col>
                <v-col cols="6">
                  <v-skeleton-loader type="image" />
                </v-col>
              </v-row>
            </template>

            <template v-else>
              <v-row dense>
                <v-col
                  v-for="col in timeseries.metricColumns"
                  :key="col.name"
                  cols="12"
                  :md="rowCols"
                >
                  <TimeseriesMetric
                    :loading="timeseries.loading"
                    :resolved="timeseries.status.isResolved()"
                    :metric="col.name"
                    :unit="col.unit"
                    :groups="selectedGroups"
                    :time="timeseries.time"
                    :event-bus="eventBus"
                  />
                </v-col>
              </v-row>

              <v-row dense>
                <v-col>
                  <TimeseriesGroupsTable
                    :uql="uql"
                    :axios-params="axiosParams"
                    :order="order"
                    :groups="timeseries.groups"
                    :grouping-columns="timeseries.groupingColumns"
                    :metric-columns="timeseries.metricColumns"
                    :events-mode="systems.isEvent"
                    @current-items="setPageGroups($event)"
                    @hover:item="eventBus.emit('hover', $event)"
                  />
                </v-col>
              </v-row>
            </template>
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useOrder } from '@/use/order'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useTimeseries, TimeseriesGroup } from '@/tracing/use-timeseries'

// Components
import TimeseriesMetric from '@/tracing/TimeseriesMetric.vue'
import TimeseriesGroupsTable from '@/tracing/TimeseriesGroupsTable.vue'

// Utilities
import { eChart as colorScheme } from '@/util/colorscheme'
import { EventBus } from '@/models/eventbus'

export default defineComponent({
  name: 'TracingTimeseries',
  components: { TimeseriesMetric, TimeseriesGroupsTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()
    const eventBus = new EventBus()
    const pageGroups = shallowRef<TimeseriesGroup[]>([])

    const order = useOrder({ desc: true })
    order.syncQueryParams()

    const timeseries = useTimeseries(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/timeseries`,
        params: props.axiosParams,
      }
    })

    const rowCols = computed(() => {
      return timeseries.metricColumns.length > 1 ? 6 : 12
    })

    const selectedGroups = computed(() => {
      return pageGroups.value.filter((group) => group._selected || group._hovered)
    })

    watch(
      () => timeseries.metricColumns,
      (columns) => {
        if (!order.column && columns.length) {
          order.column = '_avg_' + columns[0].name
        }
      },
      { flush: 'pre' },
    )

    watch(
      () => timeseries.queryInfo,
      (queryInfo) => {
        if (queryInfo) {
          props.uql.setQueryInfo(queryInfo)
        }
      },
    )

    function setPageGroups(groups: TimeseriesGroup[]) {
      groups.map((group, index) => {
        group._color = colorScheme[index % colorScheme.length]
      })
      pageGroups.value = groups
    }

    return {
      eventBus,
      rowCols,

      timeseries,

      order,
      pageGroups,
      selectedGroups,
      setPageGroups,
    }
  },
})
</script>

<style lang="scss" scoped></style>
