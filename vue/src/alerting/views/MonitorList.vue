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
          <MonitorNewMenu />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet rounded="lg" outlined class="mb-4">
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

    const metrics = useMetrics()
    const monitors = useMonitors()

    return {
      metrics,
      monitors,
    }
  },
})
</script>

<style lang="scss" scoped></style>
