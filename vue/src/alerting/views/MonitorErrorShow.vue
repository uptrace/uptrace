<template>
  <div class="container--fixed-md">
    <PageToolbar>
      <v-breadcrumbs large :items="breadcrumbs" divider=">"></v-breadcrumbs>
      <v-spacer />
      <DateRangePicker :date-range="dateRange" :range-days="1" sync-query />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-4">
      <v-skeleton-loader v-if="!monitor.data" type="card" height="800"></v-skeleton-loader>

      <template v-else>
        <v-row align="end" class="mb-2 px-4 text-subtitle-2 text-center">
          <v-col cols="auto">
            <div class="grey--text font-weight-regular">Updated at</div>
            <XDate v-if="monitor.data.updatedAt" :date="monitor.data.updatedAt" format="relative" />
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
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <MonitorErrorForm
          :date-range="dateRange"
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
import { useRouter } from '@/use/router'
import { useDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'
import { useErrorMonitor, useMonitorManager, MonitorState } from '@/alerting/use-monitors'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MonitorErrorForm from '@/alerting/MonitorErrorForm.vue'
import MonitorStateAvatar from '@/alerting/MonitorStateAvatar.vue'

export default defineComponent({
  name: 'MonitorErrorShow',
  components: {
    DateRangePicker,
    MonitorErrorForm,
    MonitorStateAvatar,
  },

  setup() {
    useTitle('Metrics Monitor')
    const { router } = useRouter()
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

      bs.push({ text: 'Edit error monitor' })

      return bs
    })

    const monitor = useErrorMonitor()
    const monitorMan = useMonitorManager()

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
