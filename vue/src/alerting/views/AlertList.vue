<template>
  <div>
    <PageToolbar :fluid="$vuetify.breakpoint.lgAndDown">
      <v-toolbar-title>Alerts</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn />
    </PageToolbar>

    <v-container :fluid="!$vuetify.breakpoint.xlOnly">
      <v-row>
        <v-col cols="4" md="3">
          <AlertsSidebar :faceted-search="facetedSearch" :facets="alerts.facets" />
        </v-col>

        <v-col cols="8" md="9">
          <v-simple-table v-if="alerts.items.length" class="border-bottom">
            <tbody>
              <tr>
                <td>
                  <AlertSelection
                    v-if="alerts.items.length"
                    :selection="selection"
                    @change="alerts.reload()"
                  />
                </td>
                <td class="d-flex align-center justify-end">
                  <AlertOrderPicker
                    v-if="alerts.items.length"
                    v-model="alerts.order.column"
                    style="max-width: 200px"
                  />
                </td>
              </tr>
            </tbody>
          </v-simple-table>

          <AlertsTable
            :loading="alerts.loading"
            :alerts="alerts.items"
            :order="alerts.order"
            :pager="alerts.pager"
            @click:alert="showAlert($event)"
            @click:chip="facetedSearch.select"
          >
            <template #action="{ alert }">
              <v-checkbox
                v-model="selection.alertIds"
                :value="alert.id"
                multiple
                :ripple="false"
                @click.stop
              ></v-checkbox>
            </template>
          </AlertsTable>

          <XPagination :pager="alerts.pager" class="mt-4" />
        </v-col>
      </v-row>
    </v-container>

    <v-dialog v-model="dialog" max-width="1200" content-class="overflow-y-scroll dialog--top">
      <v-sheet v-if="dialog">
        <AlertCard
          :loading="alerts.loading"
          :alert-id="activeAlertId"
          :fluid="$vuetify.breakpoint.mdAndDown"
          @change="alerts.reload"
        />
      </v-sheet>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRouteQuery } from '@/use/router'
import { useForceReload } from '@/use/force-reload'
import { useFacetedSearch } from '@/use/faceted-search'
import { useAlerts, useAlertSelection, Alert } from '@/alerting/use-alerts'

// Components
import ForceReloadBtn from '@/components/date/ForceReloadBtn.vue'
import AlertsSidebar from '@/alerting/AlertsSidebar.vue'
import AlertSelection from '@/alerting/AlertSelection.vue'
import AlertOrderPicker from '@/alerting/AlertOrderPicker.vue'
import AlertsTable from '@/alerting/AlertsTable.vue'
import AlertCard from '@/alerting/AlertCard.vue'

export default defineComponent({
  name: 'AlertList',
  components: {
    ForceReloadBtn,
    AlertsSidebar,
    AlertSelection,
    AlertOrderPicker,
    AlertsTable,
    AlertCard,
  },

  setup() {
    useTitle('Alerts')

    const dialog = shallowRef(false)
    const activeAlertId = shallowRef<string>()

    const { forceReloadParams } = useForceReload()

    const facetedSearch = useFacetedSearch()
    const alerts = useAlerts(
      computed(() => {
        const params: Record<string, any> = {
          ...forceReloadParams.value,
          ...facetedSearch.axiosParams,
        }

        return params
      }),
    )
    const selection = useAlertSelection(
      computed(() => {
        return alerts.items
      }),
    )

    useRouteQuery().sync({
      fromQuery(queryParams) {
        if (Object.keys(queryParams).length === 0) {
          facetedSearch.selected['state'] = ['open']
        }
      },
    })

    function showAlert(alert: Alert) {
      activeAlertId.value = alert.id
      dialog.value = true
    }

    return {
      dialog,
      activeAlertId,
      showAlert,

      facetedSearch,
      alerts,
      selection,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
