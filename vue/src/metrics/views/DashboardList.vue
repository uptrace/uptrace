<template>
  <div class="container--fixed-x-lg">
    <v-container :fluid="!$vuetify.breakpoint.lgAndDown">
      <v-row>
        <v-col>
          <DashboardNewMenu :loading:="dashboards.loading" @created="dashboards.reload()" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet rounded="lg" outlined class="mb-4">
            <v-toolbar flat color="bg--primary">
              <v-toolbar-title>Dashboards</v-toolbar-title>

              <v-col cols="auto">
                <v-text-field
                  v-model="searchInput"
                  placeholder="Search dashboards..."
                  prepend-inner-icon="mdi-magnify"
                  clearable
                  outlined
                  dense
                  hide-details="auto"
                  style="max-width: 350px"
                />
              </v-col>

              <v-spacer />
              <div v-if="dashboards.items.length" class="text-body-2">
                <NumValue
                  :value="dashboards.items.length"
                  format="verbose"
                  class="font-weight-bold"
                />
                <span> dashboards</span>
              </div>
            </v-toolbar>

            <span v-if="false">
              <v-checkbox
                v-model="dashboards.pinnedFilter"
                hide-details="auto"
                class="d-inline-block"
              />show pinned dashboards
            </span>

            <div class="pa-4">
              <DashboardsTable
                :dashboards="dashboards.items"
                :loading="dashboards.loading"
                :order="dashboards.order"
                @change="dashboards.reload"
              >
              </DashboardsTable>
            </div>
          </v-sheet>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'
import { refDebounced } from '@/use/ref-debounced'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { useProject } from '@/org/use-projects'
import { useDashboards } from '@/metrics/use-dashboards'

// Components
import DashboardNewMenu from '@/metrics/DashboardNewMenu.vue'
import DashboardsTable from '@/metrics/DashboardsTable.vue'

export default defineComponent({
  name: 'DashboardList',
  components: {
    DashboardNewMenu,
    DashboardsTable,
  },

  setup(props) {
    const searchInput = shallowRef('')
    const debouncedSearchInput = refDebounced(searchInput, 600)
    const pinnedFilter = shallowRef(false)

    const project = useProject()
    const dashboards = useDashboards(() => {
      const params: Record<string, any> = {}

      if (debouncedSearchInput.value) {
        params.q = debouncedSearchInput.value
      }

      if (pinnedFilter.value) {
        params.pinned = true
      }

      return params
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('sort_by', 'name')
        queryParams.setDefault('sort_desc', false)

        dashboards.order.parseQueryParams(queryParams)

        searchInput.value = queryParams.string('q')
        debouncedSearchInput.flush()
      },
      toQuery() {
        const queryParams: Record<string, any> = {
          q: debouncedSearchInput.value,
          ...dashboards.order.queryParams(),
        }

        return queryParams
      },
    })

    return {
      project,
      dashboards,
      searchInput,
    }
  },
})
</script>

<style lang="scss" scoped></style>
