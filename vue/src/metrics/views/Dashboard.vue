<template>
  <XPlaceholder>
    <template v-if="dashboards.isEmpty" #placeholder>
      <v-container class="fill-height">
        <v-row align="center" justify="center">
          <v-col cols="8" md="6" lg="5">
            <DashNewForm @create="onCreateDash" />
          </v-col>
        </v-row>
      </v-container>
    </template>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown">
      <v-progress-linear v-if="dashboard.loading" top absolute indeterminate></v-progress-linear>

      <v-row align="center">
        <v-col cols="auto">
          <DashPicker :dashboards="dashboards" />
        </v-col>
        <v-col v-if="dashboard.data" cols="auto">
          <DashMenu :dashboards="dashboards" :dashboard="dashboard" />
        </v-col>
        <v-col v-if="dashboard.data" cols="auto">
          <div v-if="dashboard.isTemplate">
            <DashLockedIcon />
            <DashCloneBtn :dashboard="dashboard" class="ml-2" @update:clone="dashboards.reload()" />
          </div>
          <DashToggleView v-else :dashboard="dashboard" @input:view="onChangeView" />
        </v-col>
        <v-spacer />
        <DateRangePicker :date-range="dateRange" :range-days="90" sync-query />
      </v-row>

      <v-row v-if="!dashboard.status.hasData()">
        <v-col v-for="i in 6" :key="i" cols="12" md="6">
          <v-skeleton-loader type="card" height="300px"></v-skeleton-loader>
        </v-col>
      </v-row>

      <v-row v-if="dashboard.data" :key="dashboard.data.id">
        <v-col>
          <DashTable
            v-if="dashboard.data.isTable"
            :date-range="dateRange"
            :metrics="metrics"
            :dashboard="dashboard"
            :editing="editing"
          >
          </DashTable>

          <DashGrid
            v-else
            :date-range="dateRange"
            :metrics="metrics"
            :dashboard="dashboard"
            :base-query.sync="dashboard.data.baseQuery"
          >
          </DashGrid>
        </v-col>
      </v-row>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouter, useRouteQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseMetrics } from '@/metrics/use-metrics'
import { useDashboards, useDashboard, Dashboard } from '@/metrics/use-dashboards'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import DashPicker from '@/metrics/DashPicker.vue'
import DashMenu from '@/metrics/DashMenu.vue'
import DashLockedIcon from '@/metrics/DashLockedIcon.vue'
import DashCloneBtn from '@/metrics/DashCloneBtn.vue'
import DashToggleView from '@/metrics/DashToggleView.vue'
import DashNewForm from '@/metrics/DashNewForm.vue'
import DashGrid from '@/metrics/DashGrid.vue'
import DashTable from '@/metrics/DashTable.vue'

export default defineComponent({
  name: 'MetricsDashboard',
  components: {
    DateRangePicker,
    DashPicker,
    DashMenu,
    DashLockedIcon,
    DashCloneBtn,
    DashToggleView,
    DashNewForm,
    DashGrid,
    DashTable,
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
  },

  setup() {
    useTitle('Metrics')
    const { router } = useRouter()

    const dashboards = useDashboards()
    const dashboard = useDashboard()

    const editing = shallowRef(false)

    useRouteQuery().onRouteUpdated(() => {
      editing.value = false
    })

    function onCreateDash(dash: Dashboard) {
      router.replace({ name: 'MetricsDashShow', params: { dashId: dash.id } })
      dashboards.reload()
    }

    function onChangeView() {
      editing.value = true
    }

    return {
      dashboards,
      dashboard,

      editing,

      onCreateDash,
      onChangeView,
    }
  },
})
</script>

<style lang="scss" scoped></style>
