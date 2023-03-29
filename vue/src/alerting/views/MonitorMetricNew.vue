<template>
  <div class="container--fixed-lg">
    <PageToolbar :fluid="$vuetify.breakpoint.mdAndDown">
      <v-breadcrumbs :items="breadcrumbs" divider=">" large class="pl-0"></v-breadcrumbs>
      <v-spacer />
      <DateRangePicker :date-range="dateRange" :range-days="90" />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-4">
      <v-row>
        <v-col>
          <MonitorMetricForm
            :date-range="dateRange"
            :metrics="metrics"
            :monitor="monitor"
            :column-map="columnMap"
            @click:save="onSave"
            @click:cancel="onCancel"
          />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, reactive, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouterOnly, useRoute } from '@/use/router'
import { useDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'
import { useMetrics } from '@/metrics/use-metrics'
import { emptyMetricMonitor, EmptyMetricMonitor } from '@/alerting/use-monitors'
import { MetricColumn } from '@/metrics/types'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MonitorMetricForm from '@/alerting/MonitorMetricForm.vue'

export default defineComponent({
  name: 'MonitorMetricNew',
  components: {
    DateRangePicker,
    MonitorMetricForm,
  },

  setup(_props, ctx) {
    useTitle('New Metrics Monitor')
    const router = useRouterOnly()
    const route = useRoute()
    const dateRange = useDateRange()
    const project = useProject()

    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: project.data?.name ?? 'Project',
        to: {
          name: 'ProjectShow',
        },
        exact: true,
      })

      bs.push({
        text: 'Monitors',
        to: {
          name: 'MonitorList',
        },
        exact: true,
      })

      bs.push({ text: 'New metric monitor' })

      return bs
    })

    const metrics = useMetrics()
    const monitor = reactive(emptyMetricMonitor())
    const columnMap = ref<Record<string, MetricColumn>>({})
    initMonitorFromQuery(monitor)

    function initMonitorFromQuery(monitor: EmptyMetricMonitor) {
      const routeQuery = route.value.query
      monitor.name = asString(routeQuery.name)
      monitor.params.query = asString(routeQuery.query)

      const columns = routeQuery.columns
      if (columns && typeof columns === 'string') {
        columnMap.value = JSON.parse(columns)
      }

      const metrics = asArray(routeQuery.metric)
      if (!metrics || !metrics.length) {
        return
      }

      const aliases = asArray(routeQuery.alias)
      if (!aliases || !aliases.length) {
        return
      }

      if (metrics.length !== aliases.length) {
        return
      }

      metrics.forEach((metric, index) => {
        monitor.params.metrics.push({ name: metric, alias: aliases[index] })
      })
    }

    function onSave() {
      redirectToMonitors()
    }
    function onCancel() {
      redirectToMonitors()
    }
    function redirectToMonitors() {
      router.push({ name: 'MonitorList' })
    }

    return {
      dateRange,
      breadcrumbs,

      metrics,
      monitor,
      columnMap,

      onSave,
      onCancel,
    }
  },
})

function asString(v: any): string {
  if (typeof v === 'string') {
    return v
  }
  return ''
}

function asArray(v: any): string[] {
  if (!v) {
    return []
  }
  if (Array.isArray(v)) {
    return v
  }
  return [v]
}
</script>

<style lang="scss" scoped></style>
