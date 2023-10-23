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
      </v-container>

      <div class="border">
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="py-0">
          <v-row align="center" no-gutters>
            <v-col v-if="$route.params.dashId" cols="auto">
              <v-tabs>
                <v-tab :to="{ name: 'DashboardTable' }" exact-path>Table</v-tab>
                <v-tab :to="{ name: 'DashboardGrid' }" exact-path>Grid</v-tab>
                <v-tab :to="{ name: 'DashboardYaml' }" exact-path>YAML</v-tab>
                <v-tab :to="{ name: 'DashboardHelp' }" exact-path>Help</v-tab>
              </v-tabs>
            </v-col>
          </v-row>
        </v-container>
      </div>

      <router-view
        v-if="dashboard.data"
        name="tab"
        :date-range="dateRange"
        :dashboard="dashboard.data"
        :grid="dashboard.grid"
        :grid-query="dashboard.data.gridQuery"
        editable
        @change="dashboard.reload"
      />
      <v-container v-else :fluid="$vuetify.breakpoint.lgAndDown">
        <v-skeleton-loader type="card,table"></v-skeleton-loader>
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouterOnly, useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useAnnotations } from '@/org/use-annotations'
import { useDashboards, useDashboard } from '@/metrics/use-dashboards'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import DashPicker from '@/metrics/DashPicker.vue'
import DashMenu from '@/metrics/DashMenu.vue'
import DashPinBtn from '@/metrics/DashPinBtn.vue'
import DashNewForm from '@/metrics/DashNewForm.vue'

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
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    const router = useRouterOnly()
    const route = useRoute()

    useAnnotations(() => {
      return {
        ...props.dateRange.axiosParams(),
      }
    })

    const dashboards = useDashboards()
    const dashboard = useDashboard()
    useTitle(computed(() => `${dashboard.data?.name ?? 'Dashboard'} | Dashboard`))

    watch(
      () => dashboard.data,
      (data) => {
        if (!dashboard.data) {
          return
        }
        if (route.value.name !== 'DashboardShow') {
          return
        }

        if (dashboard.data.tableMetrics.length && dashboard.data.tableQuery) {
          router.push({ name: 'DashboardTable' })
          return
        }

        if (dashboard.grid.length) {
          router.push({ name: 'DashboardGrid' })
          return
        }

        router.push({ name: 'DashboardTable' })
      },
    )

    function onCreateDash(dash: Dashboard) {
      dashboards.reload().then(() => {
        router.replace({ name: 'DashboardShow', params: { dashId: String(dash.id) } })
      })
    }

    function onCloneDash(dash: Dashboard) {
      dashboards.reload().then(() => {
        router.replace({ name: 'DashboardShow', params: { dashId: String(dash.id) } })
      })
    }

    function onPinDash() {
      dashboards.reload()
      dashboard.reload()
    }

    return {
      DashKind,

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
