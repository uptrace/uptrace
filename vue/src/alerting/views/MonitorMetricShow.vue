<template>
  <div class="container--fixed-md">
    <PageToolbar>
      <v-breadcrumbs large :items="breadcrumbs" divider=">"></v-breadcrumbs>
      <v-spacer />
      <DateRangePicker :date-range="dateRange" :range-days="1" />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-4">
      <v-skeleton-loader v-if="!monitor.data" type="card" height="800"></v-skeleton-loader>

      <template v-else>
        <v-row align="end" class="mb-2 px-4 text-subtitle-2 text-center">
          <v-col cols="auto">
            <div class="grey--text font-weight-regular">Last check at</div>
            <DateValue
              v-if="monitor.data.updatedAt"
              :value="monitor.data.updatedAt"
              format="relative"
            />
            <div v-else>None</div>
          </v-col>
          <v-col cols="auto">
            <div class="grey--text font-weight-regular">State</div>
            <div><MonitorStateAvatar :state="monitor.data.state" /></div>
          </v-col>
          <v-col cols="auto">
            <v-btn
              v-if="monitor.data.state != MonitorState.Paused"
              :loading="monitorMan.pending"
              depressed
              small
              title="Pause monitor"
              @click="pauseMonitor(monitor)"
            >
              <v-icon left>mdi-pause</v-icon>
              Pause monitor
            </v-btn>
            <v-btn
              v-else
              :loading="monitorMan.pending"
              depressed
              small
              title="Resume monitor"
              @click="activateMonitor(monitor)"
            >
              <v-icon left>mdi-play</v-icon>
              Resume monitor
            </v-btn>
          </v-col>
          <v-col cols="auto">
            <v-btn
              depressed
              small
              :to="{
                name: 'AlertList',
                query: {
                  q: 'monitor:' + monitor.data.id,
                  state: null,
                },
              }"
            >
              View alerts
            </v-btn>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <MonitorMetricForm
          :date-range="dateRange"
          :metrics="metrics"
          :monitor="reactive(monitor.data)"
          @click:save="onSave"
          @click:cancel="onCancel"
        />
      </template>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, reactive } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouterOnly, useSyncQueryParams } from '@/use/router'
import { useDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'
import { useMetrics } from '@/metrics/use-metrics'
import { useMetricMonitor, useMonitorManager, MonitorState } from '@/alerting/use-monitors'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MonitorMetricForm from '@/alerting/MonitorMetricForm.vue'
import MonitorStateAvatar from '@/alerting/MonitorStateAvatar.vue'

export default defineComponent({
  name: 'MonitorMetricShow',
  components: {
    DateRangePicker,
    MonitorMetricForm,
    MonitorStateAvatar,
  },

  setup() {
    useTitle('Metrics Monitor')
    const router = useRouterOnly()
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

      bs.push({ text: 'Edit metric monitor' })

      return bs
    })

    const metrics = useMetrics()
    const monitor = useMetricMonitor()
    const monitorMan = useMonitorManager()

    useSyncQueryParams({
      fromQuery(queryParams) {
        dateRange.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...dateRange.queryParams(),
        }
      },
    })

    function onSave() {
      redirectToMonitors()
    }
    function onCancel() {
      redirectToMonitors()
    }
    function redirectToMonitors() {
      router.push({ name: 'MonitorList' })
    }

    function activateMonitor() {
      if (!monitor.data) {
        return
      }
      monitorMan.activate(monitor.data).then(() => {
        monitor.reload()
      })
    }

    function pauseMonitor() {
      if (!monitor.data) {
        return
      }
      monitorMan.pause(monitor.data).then(() => {
        monitor.reload()
      })
    }

    return {
      MonitorState,

      dateRange,
      breadcrumbs,

      metrics,
      monitor,

      onSave,
      onCancel,

      monitorMan,
      activateMonitor,
      pauseMonitor,
      reactive,
    }
  },
})
</script>

<style lang="scss" scoped></style>
