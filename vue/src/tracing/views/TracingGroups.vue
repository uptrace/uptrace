<template>
  <div>
    <v-row v-if="groups.errorCode === 'invalid_query'">
      <v-col>
        <v-banner>
          <v-icon slot="icon" color="error" size="36">mdi-alert-circle</v-icon>
          <div class="subtitle-1 text--secondary">
            {{ groups.errorMessage }}<br />
            This is a bug. Please report in on
            <a href="https://github.com/uptrace/uptrace" target="_blank">GitHub</a> including the
            error message and the query.
          </div>
        </v-banner>

        <PrismCode v-if="groups.backendQuery" :code="groups.backendQuery" language="sql" />
      </v-col>
    </v-row>

    <v-row v-else>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>Groups</span>
            </v-toolbar-title>

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <span v-if="groups.hasMore">more than </span>
              <strong><XNum :value="numGroup" verbose /></strong> groups
            </div>
          </v-toolbar>

          <v-container fluid>
            <GroupsList
              :date-range="dateRange"
              :systems="systems.activeSystems"
              :uql="uql"
              :loading="groups.loading"
              :groups="groups.items"
              :columns="groups.columns"
              :plottable-columns="groups.plottableColumns"
              :plotted-columns="plottedColumns"
              show-plotted-column-items
              :order="groups.order"
              :events-mode="systems.isEvent"
              :show-system="showSystem"
              :axios-params="groups.axiosParams"
              @update:num-group="numGroup = $event"
            />
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, watchEffect, PropType } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'

// Components
import GroupsList from '@/tracing/GroupsList.vue'

// Utilities
import { isGroupSystem } from '@/models/otel'

export default defineComponent({
  name: 'TracingGroups',
  components: { GroupsList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    props.dateRange.roundUp()

    const groups = useGroups(() => {
      return props.axiosParams
    })
    groups.order.syncQueryParams()
    const numGroup = shallowRef(0)

    const showSystem = computed(() => {
      const systems = props.systems.activeSystems
      if (systems.length > 1) {
        return true
      }
      if (systems.length === 1) {
        return isGroupSystem(systems[0])
      }
      return false
    })

    const plottedColumns = shallowRef<string[]>()
    watchEffect(() => {
      if (!groups.plottableColumns.length) {
        plottedColumns.value = undefined
        return
      }

      if (!plottedColumns.value) {
        plottedColumns.value = groups.plottableColumns.slice(0, 1).map((col) => col.name)
        return
      }

      plottedColumns.value = plottedColumns.value.filter((colName) => {
        return groups.plottableColumns.findIndex((item) => item.name === colName) >= 0
      })
    })
    useRouteQuery().sync({
      fromQuery(params) {
        if (Array.isArray(params.columns)) {
          plottedColumns.value = params.columns
        } else if (params.columns) {
          plottedColumns.value = [params.columns]
        } else if (params.column) {
          plottedColumns.value = [params.column]
        }
      },
      toQuery() {
        return {
          columns: plottedColumns.value,
        }
      },
    })

    watch(
      () => groups.queryInfo,
      (queryInfo) => {
        if (queryInfo) {
          props.uql.setQueryInfo(queryInfo)
        }
      },
    )

    return {
      groups,

      numGroup,
      showSystem,
      plottedColumns,
    }
  },
})
</script>

<style lang="scss" scoped></style>
