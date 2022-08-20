<template>
  <v-col cols="12" md="6">
    <v-card outlined rounded="lg">
      <v-toolbar color="light-blue lighten-5" flat dense>
        <v-toolbar-items class="mr-2">
          <v-btn
            v-if="!dashboard.isTemplate"
            icon
            title="Drag and drop to change position"
            class="draggable-handle"
            style="cursor: move"
          >
            <v-icon>mdi-drag</v-icon>
          </v-btn>
        </v-toolbar-items>

        <v-toolbar-title>
          <span :class="{ 'red--text': timeseries.hasError }">{{ dashEntry.name }}</span>
          <v-icon v-if="timeseries.hasError" color="error" title="The query has errors" class="ml-2"
            >mdi-alert-circle-outline</v-icon
          >

          <v-tooltip v-if="dashEntry.description" bottom>
            <template #activator="{ on, attrs }">
              <v-icon class="ml-2" v-bind="attrs" v-on="on">mdi-information-outline</v-icon>
            </template>
            <span>{{ dashEntry.description }}</span>
          </v-tooltip>
        </v-toolbar-title>

        <v-spacer />

        <v-toolbar-items>
          <v-menu offset-y>
            <template #activator="{ on: onMenu, attrs }">
              <v-btn :loading="dashEntryMan.pending" icon v-bind="attrs" v-on="onMenu">
                <v-icon>mdi-dots-vertical</v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item @click="dialog = true">
                <v-list-item-title>Edit</v-list-item-title>
              </v-list-item>

              <v-list-item v-if="!dashboard.isTemplate" @click="del">
                <v-list-item-title>Delete</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-toolbar-items>
      </v-toolbar>

      <v-card-text class="pa-2">
        <v-row>
          <v-col>
            <MetricChart
              :loading="timeseries.loading"
              :resolved="timeseries.status.isResolved()"
              :chart-type="dashEntry.chartType"
              :column-map="dashEntry.columnMap"
              :timeseries="timeseries.items"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <MetricSummary :timeseries="timeseries.items" />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <DashGridEntryFormDialog
      :date-range="dateRange"
      :value="dialog"
      :metrics="metrics"
      :dashboard="dashboard"
      :dash-entry="dashEntry"
      :timeseries="timeseries"
      @change="
        $emit('change', $event)
        dialog = false
      "
    >
    </DashGridEntryFormDialog>
  </v-col>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseMetrics } from '@/metrics/use-metrics'
import { useDashEntryManager, UseDashboard, DashEntry } from '@/metrics/use-dashboards'
import { useTimeseries } from '@/metrics/use-query'

// Components
import MetricChart from '@/metrics/MetricChart.vue'
import MetricSummary from '@/metrics/MetricSummary.vue'
import DashGridEntryFormDialog from '@/metrics/DashGridEntryFormDialog.vue'

export default defineComponent({
  name: 'DashEntryCol',
  components: {
    MetricSummary,
    MetricChart,
    DashGridEntryFormDialog,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
    dashEntry: {
      type: Object as PropType<DashEntry>,
      required: true,
    },
    baseQuery: {
      type: String,
      required: true,
    },
    editing: {
      type: Boolean,
      required: true,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const dialog = shallowRef(false)
    const dashEntryMan = useDashEntryManager()

    const timeseries = useTimeseries(() => {
      if (!props.dashboard.status.isResolved()) {
        return undefined
      }
      if (!props.dashEntry.metrics || !props.dashEntry.metrics.length || !props.dashEntry.query) {
        return undefined
      }

      const metrics = []
      const aliases = []

      for (let metric of props.dashEntry.metrics) {
        metrics.push(metric.name)
        aliases.push(metric.alias)
      }

      const { projectId } = route.value.params
      const req = {
        url: `/api/v1/metrics/${projectId}/timeseries`,
        params: {
          ...props.dateRange.axiosParams(),
          metrics,
          aliases,
          base_query: props.baseQuery,
          query: props.dashEntry.query,
        },
      }
      return req
    })

    watch(
      () => timeseries.baseQuery,
      (baseQuery) => {
        if (baseQuery) {
          ctx.emit('input:base-query', baseQuery)
        }
      },
    )

    watch(
      () => props.editing,
      (editing) => {
        if (editing) {
          dialog.value = editing
        }
      },
      { immediate: true },
    )

    function del() {
      dashEntryMan.del(props.dashEntry).then(() => {
        ctx.emit('change')
      })
    }

    return {
      dialog,

      timeseries,
      dashEntryMan,
      del,
    }
  },
})
</script>

<style lang="scss" scoped></style>
