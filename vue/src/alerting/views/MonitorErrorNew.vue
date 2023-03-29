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
          <MonitorErrorForm
            :date-range="dateRange"
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
import { useRouterOnly } from '@/use/router'
import { useDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'
import { createEmptyErrorMonitor } from '@/alerting/use-monitors'
import { MetricColumn } from '@/metrics/types'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MonitorErrorForm from '@/alerting/MonitorErrorForm.vue'

export default defineComponent({
  name: 'MonitorErrorNew',
  components: {
    DateRangePicker,
    MonitorErrorForm,
  },

  setup(_props, ctx) {
    useTitle('New Metrics Monitor')
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

      bs.push({ text: 'New error monitor' })

      return bs
    })

    const monitor = reactive(createEmptyErrorMonitor())
    const columnMap = ref<Record<string, MetricColumn>>({})

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

      monitor,
      columnMap,

      onSave,
      onCancel,
    }
  },
})
</script>

<style lang="scss" scoped></style>
