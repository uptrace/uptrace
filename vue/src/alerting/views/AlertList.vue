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
                  <AlertSelection :selection="selection" @change="alerts.reload()" />
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
            <template #prepend-column="{ alert }">
              <td class="pr-0" @click.stop="selection.toggle(alert)">
                <v-checkbox
                  v-model="selection.alertIds"
                  :value="alert.id"
                  multiple
                  @click.stop.prevent
                ></v-checkbox>
              </td>
            </template>
          </AlertsTable>

          <XPagination :pager="pager" class="mt-4" />
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
import { defineComponent, shallowRef, computed, watch } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useSyncQueryParams } from '@/use/router'
import { useForceReload } from '@/use/force-reload'
import { usePager } from '@/use/pager'
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
    const activeAlertId = shallowRef<number>()

    const { forceReloadParams } = useForceReload()

    const pager = usePager()
    const facetedSearch = useFacetedSearch()
    const alerts = useAlerts(
      computed(() => {
        const params: Record<string, any> = {
          ...forceReloadParams.value,
          ...facetedSearch.axiosParams(),
        }

        return params
      }),
    )
    const pageAlerts = computed(() => {
      return alerts.items.slice(pager.pos.start, pager.pos.end)
    })

    const selection = useAlertSelection(
      computed(() => {
        return alerts.items
      }),
      pageAlerts,
    )

    useSyncQueryParams({
      fromQuery(queryParams) {
        if (queryParams.empty()) {
          queryParams.setDefault('attrs.alert.status', 'open')
        }

        queryParams.setDefault('sort_by', 'updated_at')
        queryParams.setDefault('sort_desc', true)

        alerts.order.parseQueryParams(queryParams)
        facetedSearch.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...alerts.order.queryParams(),
          ...facetedSearch.queryParams(),
        }
      },
    })

    watch(
      () => alerts.items.length,
      (numItem) => {
        pager.numItem = numItem
      },
    )

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

      pager,
      pageAlerts,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
