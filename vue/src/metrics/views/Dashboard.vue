<template>
  <div>
    <v-container v-if="dashboards.isEmpty" class="fill-height">
      <v-row align="center" justify="center">
        <v-col cols="8" md="6" lg="5">
          <DashNewForm @create="onCreateDash" />
        </v-col>
      </v-row>
    </v-container>

    <template v-else>
      <v-container :fluid="$vuetify.breakpoint.lgAndDown">
        <v-progress-linear v-if="dashboard.loading" top absolute indeterminate></v-progress-linear>

        <v-row align="center">
          <v-col cols="auto">
            <DashPicker
              :loading="dashboards.loading"
              :value="dashboards.active?.id"
              :items="dashboards.items"
            />
          </v-col>
          <v-col v-if="dashboard.data" cols="auto">
            <DashMenu :dashboards="dashboards" :dashboard="dashboard" />
            <DashPinBtn v-if="dashboard.data" :dashboard="dashboard.data" @update="onPinDash" />
          </v-col>
          <v-spacer />
          <v-col cols="auto">
            <DateRangePicker :date-range="dateRange" :range-days="90" />
          </v-col>
        </v-row>

        <v-row v-if="!dashboard.status.hasData()">
          <v-col v-for="i in 6" :key="i" cols="12" md="6">
            <v-skeleton-loader type="card" height="300px"></v-skeleton-loader>
          </v-col>
        </v-row>
      </v-container>

      <div class="border">
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="py-0">
          <v-row align="center" no-gutters>
            <v-col cols="auto">
              <v-tabs v-model="activeTab">
                <v-tab href="#table">Table</v-tab>
                <v-tab href="#grid">Grid</v-tab>
                <v-tab href="#yaml">YAML</v-tab>
                <v-tab href="#help">Help</v-tab>
              </v-tabs>
            </v-col>
          </v-row>
        </v-container>
      </div>

      <v-tabs-items v-if="dashboard.data" v-model="activeTab">
        <v-tab-item :value="DashKind.Table">
          <DashTable
            :date-range="dateRange"
            :dashboard="dashboard.data"
            :grid="dashboard.grid"
            editable
            @change="dashboard.reload"
          >
          </DashTable>
        </v-tab-item>
        <v-tab-item :value="DashKind.Grid">
          <DashGrid
            :date-range="dateRange"
            :dashboard="dashboard.data"
            :grid="dashboard.grid"
            :grid-query="dashboard.data.gridQuery"
            editable
            @change="dashboard.reload"
          >
          </DashGrid>
        </v-tab-item>
        <v-tab-item value="yaml">
          <DashYaml :yaml-url="dashboard.yamlUrl" @change="dashboard.reload" />
        </v-tab-item>
        <v-tab-item value="help"><DashHelp /></v-tab-item>
      </v-tabs-items>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouter, useRouteQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useAnnotations } from '@/org/use-annotations'
import { useDashboards, useDashboard } from '@/metrics/use-dashboards'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import DashPicker from '@/metrics/DashPicker.vue'
import DashMenu from '@/metrics/DashMenu.vue'
import DashPinBtn from '@/metrics/DashPinBtn.vue'
import DashNewForm from '@/metrics/DashNewForm.vue'
import DashGrid from '@/metrics/DashGrid.vue'
import DashTable from '@/metrics/DashTable.vue'
import DashYaml from '@/metrics/DashYaml.vue'
import DashHelp from '@/metrics/DashHelp.vue'

// Types
import { Dashboard, DashKind } from '@/metrics/types'

export default defineComponent({
  name: 'MetricsDashboard',
  components: {
    DateRangePicker,
    DashPicker,
    DashMenu,
    DashPinBtn,
    DashNewForm,
    DashGrid,
    DashTable,
    DashYaml,
    DashHelp,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Metrics')
    props.dateRange.syncQueryParams()

    useAnnotations(() => {
      return {
        ...props.dateRange.axiosParams(),
      }
    })

    const { router } = useRouter()

    const dashboards = useDashboards()
    const dashboard = useDashboard()

    const activeTab = shallowRef(DashKind.Table)
    watch(
      () => dashboard.data?.id,
      () => {
        const dash = dashboard.data
        if (!dash) {
          return
        }
        switch (activeTab.value) {
          case DashKind.Table:
            if (!dash.tableMetrics.length || !dash.tableQuery) {
              if (dashboard.grid.length) {
                activeTab.value = DashKind.Grid
              }
            }
            return
          case DashKind.Grid:
            if (!dashboard.grid.length) {
              if (dash.tableMetrics.length || dash.tableQuery) {
                activeTab.value = DashKind.Table
              }
            }
            return
        }
      },
    )

    useRouteQuery().sync({
      fromQuery(params) {
        if (params.tab) {
          activeTab.value = params.tab
        }
      },
      toQuery() {
        if (activeTab.value) {
          return { tab: activeTab.value }
        }
        return {}
      },
    })

    function onCreateDash(dash: Dashboard) {
      dashboards.reload().then(() => {
        router.replace({ name: 'MetricsDashShow', params: { dashId: String(dash.id) } })
      })
    }

    function onCloneDash(dash: Dashboard) {
      dashboards.reload().then(() => {
        router.replace({ name: 'MetricsDashShow', params: { dashId: String(dash.id) } })
      })
    }

    function onPinDash() {
      dashboards.reload()
      dashboard.reload()
    }

    return {
      DashKind,
      activeTab,

      dashboards,
      dashboard,

      onCreateDash,
      onCloneDash,
      onPinDash,
    }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
