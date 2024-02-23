<template>
  <div>
    <v-container
      v-if="dashboards.isEmpty"
      class="fill-height"
      style="min-height: calc(100vh - 240px)"
    >
      <v-row align="center" justify="center">
        <v-col cols="8" md="6" lg="5">
          <DashboardForm @saved="dashboards.reload()" />
        </v-col>
      </v-row>
    </v-container>

    <template v-else>
      <v-container :fluid="$vuetify.breakpoint.lgAndDown">
        <v-progress-linear v-if="dashboard.loading" top absolute indeterminate></v-progress-linear>

        <v-row align="center" dense>
          <v-col cols="auto">
            <DashPicker :loading="dashboards.loading" :items="dashboards.items" />
          </v-col>
          <v-col v-if="dashboard.data" cols="auto">
            <DashboardMenu
              :dashboard="dashboard.data"
              @created="onCreateDashboard($event)"
              @updated="
                dashboard.reload()
                dashboards.reload()
              "
              @deleted="$router.push({ name: 'DashboardList' })"
            />
            <DashPinBtn :dashboard="dashboard.data" @update="onPinDash" />
          </v-col>
          <v-col v-if="dashboard.data" cols="auto">
            <RelatedDashboardsTabs
              :dashboard="dashboard.data"
              :dashboards="dashboards.items"
              class="ml-2"
            />
          </v-col>
          <v-spacer />
        </v-row>
      </v-container>

      <div class="border">
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="py-0">
          <v-row align="center">
            <v-col v-if="$route.params.dashId" cols="auto">
              <v-tabs>
                <v-tab :to="{ name: 'DashboardTable' }" exact-path>Table</v-tab>
                <v-tab :to="{ name: 'DashboardGrid' }" exact-path>Grid</v-tab>
                <v-tab :to="{ name: 'DashboardHelp' }" exact-path>Help</v-tab>
              </v-tabs>
            </v-col>
            <v-spacer />
            <portal-target name="dashboard-actions"></portal-target>
            <v-col cols="auto">
              <DateRangePicker :date-range="dateRange" :range-days="90" />
            </v-col>
          </v-row>
        </v-container>
      </div>

      <v-card flat min-height="calc(100vh - 242px)" color="grey lighten-5">
        <router-view
          v-if="dashboard.data"
          name="tab"
          :date-range="dateRange"
          :dashboard="dashboard.data"
          :table-items="dashboard.tableItems"
          :grid-rows="dashboard.gridRows"
          :grid-metrics="dashboard.gridMetrics"
          :grid-query="dashboard.data.gridQuery"
          @change="dashboard.reload()"
        />
        <v-container v-else :fluid="$vuetify.breakpoint.lgAndDown">
          <v-skeleton-loader type="card,table"></v-skeleton-loader>
        </v-container>
      </v-card>
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
import DashboardMenu from '@/metrics/DashboardMenu.vue'
import DashPinBtn from '@/metrics/DashPinBtn.vue'
import DashboardForm from '@/metrics/DashboardForm.vue'
import RelatedDashboardsTabs from '@/metrics/RelatedDashboardsTabs.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'Dashboard',
  components: {
    DateRangePicker,
    DashPicker,
    DashboardMenu,
    DashPinBtn,
    DashboardForm,
    RelatedDashboardsTabs,
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

        if (dashboard.gridRows.length) {
          router.push({ name: 'DashboardGrid' })
          return
        }

        router.push({ name: 'DashboardTable' })
      },
    )

    function onCreateDashboard(dash: Dashboard) {
      dashboards.reload().then(() => {
        router.replace({ name: 'DashboardShow', params: { dashId: String(dash.id) } })
      })
    }

    function onPinDash() {
      dashboards.reload()
      dashboard.reload()
    }

    return {
      dashboards,
      dashboard,

      onCreateDashboard,
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
