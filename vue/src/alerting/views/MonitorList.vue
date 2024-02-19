<template>
  <div>
    <PageToolbar :fluid="$vuetify.breakpoint.lgAndDown">
      <v-toolbar-title>Monitors</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.lgAndDown">
      <v-row>
        <v-col>
          <MonitorNewMenu @create="monitors.reload()" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet rounded="lg" outlined class="mb-4">
            <v-toolbar flat dense color="light-blue lighten-5">
              <v-text-field
                v-model="monitors.searchInput"
                placeholder="Search monitors..."
                prepend-inner-icon="mdi-magnify"
                clearable
                outlined
                dense
                hide-details="auto"
                style="max-width: 350px"
              />

              <v-spacer />
              <div v-if="monitors.items.length" class="text-body-2 blue-grey--text text--darken-3">
                <NumValue
                  :value="monitors.items.length"
                  format="verbose"
                  class="font-weight-bold"
                />
                <span> monitors</span>
              </div>
            </v-toolbar>

            <div class="pa-4">
              <v-skeleton-loader
                v-if="!monitors.status.hasData()"
                type="table"
                height="600px"
              ></v-skeleton-loader>

              <template v-else>
                <MonitorStateCounts
                  :states="monitors.states"
                  @input="monitors.stateFilter = $event"
                />

                <MonitorsTable
                  :loading="monitors.loading"
                  :monitors="monitors.items"
                  :order="monitors.order"
                  @change="monitors.reload()"
                >
                </MonitorsTable>
              </template>
            </div>
          </v-sheet>

          <XPagination :pager="monitors.pager" />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { useProject } from '@/org/use-projects'
import { useTitle } from '@vueuse/core'
import { useMetrics } from '@/metrics/use-metrics'
import { useMonitors } from '@/alerting/use-monitors'

// Components
import ForceReloadBtn from '@/components/ForceReloadBtn.vue'
import MonitorNewMenu from '@/alerting/MonitorNewMenu.vue'
import MonitorsTable from '@/alerting/MonitorsTable.vue'
import MonitorStateCounts from '@/alerting/MonitorStateCounts.vue'

export default defineComponent({
  name: 'MonitorList',
  components: {
    ForceReloadBtn,
    MonitorNewMenu,
    MonitorsTable,
    MonitorStateCounts,
  },

  setup() {
    useTitle('Monitors')

    const project = useProject()
    const metrics = useMetrics()
    const monitors = useMonitors()

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('sort_by', 'updatedAt')
        queryParams.setDefault('sort_desc', true)

        monitors.parseQueryParams(queryParams)
      },
      toQuery() {
        const queryParams: Record<string, any> = {
          ...monitors.queryParams(),
        }

        return queryParams
      },
    })

    return {
      project,
      metrics,
      monitors,
    }
  },
})
</script>

<style lang="scss" scoped></style>
