<template>
  <MonitorMetricForm
    :date-range="dateRange"
    :metrics="metrics.items"
    :monitor="activeMonitor"
    :column-map="columnMap"
    @saved="onSave"
    @click:cancel="onCancel"
  >
    <template #title>
      <v-col cols="auto">
        <v-breadcrumbs :items="breadcrumbs" divider=">" large class="pl-0"></v-breadcrumbs>
      </v-col>
      <v-col cols="auto">
        <DateRangePicker :date-range="dateRange" :range-days="90" />
      </v-col>
    </template>
  </MonitorMetricForm>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, reactive, computed, onBeforeUnmount, inject, Ref } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouterOnly, useRoute } from '@/use/router'
import { provideForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'
import { useMetrics } from '@/metrics/use-metrics'
import { useMetricMonitor } from '@/alerting/use-monitors'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MonitorMetricForm from '@/alerting/MonitorMetricForm.vue'

// Misc
import { emptyMetricMonitor, Monitor, MetricMonitor } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorMetric',
  components: {
    DateRangePicker,
    MonitorMetricForm,
  },

  setup(_props, ctx) {
    const router = useRouterOnly()
    const route = useRoute()

    provideForceReload()
    const dateRange = useDateRange()
    const project = useProject()

    const header = inject<Ref<boolean>>('header')!
    const footer = inject<Ref<boolean>>('footer')!
    header.value = false
    footer.value = false
    onBeforeUnmount(() => {
      header.value = true
      footer.value = true
    })

    const metrics = useMetrics()
    const metricMonitor = useMetricMonitor()
    const activeMonitor = computed(() => {
      if (metricMonitor.data) {
        return reactive(cloneDeep(metricMonitor.data))
      }

      const monitor = reactive(emptyMetricMonitor())
      initMonitorFromRoute(monitor)
      return monitor
    })

    useTitle(
      computed(() => {
        if (activeMonitor.value) {
          return 'Edit Metric Monitor'
        }
        return 'New Metric Monitor'
      }),
    )
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

      if (activeMonitor.value) {
        bs.push({ text: 'Edit metric monitor' })
      } else {
        bs.push({ text: 'New metric monitor' })
      }

      return bs
    })

    const columnMap = computed(() => {
      const columns = route.value.query.columns
      if (columns && typeof columns === 'string') {
        return JSON.parse(columns)
      }
      return {}
    })

    function initMonitorFromRoute(monitor: MetricMonitor) {
      const routeQuery = route.value.query
      monitor.name = asString(routeQuery.name)
      monitor.params.query = asString(routeQuery.query)

      if ('time_offset' in routeQuery) {
        monitor.params.timeOffset = parseInt(asString(routeQuery.time_offset), 10)
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

    function onSave(monitor: Monitor) {
      redirectToMonitors(monitor)
    }
    function onCancel(monitor: Monitor) {
      redirectToMonitors(monitor)
    }
    function redirectToMonitors(monitor: Monitor) {
      router.push({ name: 'MonitorList', query: { q: 'monitor:' + monitor.id } })
    }

    return {
      dateRange,
      breadcrumbs,

      metrics,
      activeMonitor,
      columnMap,

      onSave,
      onCancel,
    }
  },
})

function asString(v: any): string {
  switch (typeof v) {
    case 'string':
      return v
    case 'number':
      return String(v)
    default:
      return ''
  }
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
