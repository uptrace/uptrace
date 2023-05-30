<template>
  <v-data-table
    :loading="loading"
    :headers="headers"
    :items="groups"
    item-key="_id"
    :items-per-page="15"
    hide-default-footer
    single-expand
    :sort-by.sync="order.column"
    :sort-desc.sync="order.desc"
    must-sort
    :dense="dense"
  >
    <template #item="{ item, isExpanded, expand }">
      <GroupsTableRow
        :uql="uql"
        :events-mode="eventsModeFor(item)"
        :grouping-columns="groupingColumns"
        :plain-columns="plainColumns"
        :plottable-columns="plottableColumns"
        :plotted-columns="plottedColumns"
        :axios-params="internalAxiosParams"
        :headers="headers"
        :column-map="columnMap"
        :group="item"
        :is-expanded="isExpanded"
        :expand="expand"
        @click:metrics="$emit('click:metrics', $event)"
      >
      </GroupsTableRow>
    </template>
    <template #expanded-item="{ headers, item }">
      <tr class="v-data-table__expanded v-data-table__expanded__content">
        <td :colspan="headers.length" class="pt-2 pb-4">
          <SpansList
            :events-mode="eventsModeFor(item)"
            :uql="uql"
            :axios-params="internalAxiosParams"
            :where="item._query"
          />
        </td>
      </tr>
    </template>
    <template #no-data>
      <div class="pa-4">
        <div class="mb-4">There are no matching groups.Try to change filters.</div>
        <v-btn :to="{ name: 'TracingHelp' }">
          <v-icon left>mdi-help-circle-outline</v-icon>
          <span>Help</span>
        </v-btn>
      </div>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { UseOrder } from '@/use/order'
import { Group, ColumnInfo } from '@/tracing/use-explore-spans'
import { UseUql } from '@/use/uql'

// Components
import GroupsTableRow from '@/tracing/GroupsTableRow.vue'
import SpansList from '@/tracing/SpansList.vue'

// Utilities
import { isEventSystem, AttrKey } from '@/models/otel'
import { updateColumnMap, MetricColumn } from '@/metrics/types'

export default defineComponent({
  name: 'GroupsTable',
  components: {
    GroupsTableRow,
    SpansList,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      default: undefined,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    isResolved: {
      type: Boolean,
      required: true,
    },
    groups: {
      type: Array as PropType<Group[]>,
      required: true,
    },
    columns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plottableColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plottedColumns: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      default: false,
    },
    dense: {
      type: Boolean,
      default: false,
    },
    showSystem: {
      type: Boolean,
      default: false,
    },
    axiosParams: {
      type: Object,
      required: true,
    },
  },

  setup(props) {
    const hasGroupName = computed((): boolean => {
      return hasColumn(AttrKey.displayName)
    })

    const hasSystemColumn = computed(() => {
      if (props.columns.findIndex((col) => col.name === AttrKey.spanSystem) >= 0) {
        return true
      }
      if (!props.showSystem) {
        return false
      }
      return hasColumn(AttrKey.spanSystem)
    })

    const groupingColumns = computed(() => {
      return props.columns.filter((col) => col.isGroup).map((col) => col.name)
    })

    const plainColumns = computed(() => {
      const blacklist: string[] = [AttrKey.spanSystem as string]
      if (hasGroupName.value) {
        blacklist.push(AttrKey.spanGroupId, AttrKey.displayName)
      }
      return props.columns.filter((col) => {
        if (props.plottableColumns.findIndex((item) => item.name === col.name) >= 0) {
          return false
        }
        if (blacklist.indexOf(col.name) >= 0) {
          return false
        }
        return true
      })
    })

    const headers = computed(() => {
      const headers = []

      if (hasGroupName.value) {
        headers.push({
          text: 'Group name',
          value: '_name',
          sortable: true,
          align: 'start',
        })
      }

      for (let col of plainColumns.value) {
        headers.push({
          text: shortColumnName(col.name),
          value: col.name,
          sortable: true,
          align: 'start',
        })
      }

      if (hasSystemColumn.value) {
        headers.push({
          text: 'System',
          value: AttrKey.spanSystem,
          sortable: true,
          align: 'start',
        })
      }

      for (let col of props.plottableColumns) {
        headers.push({
          text: shortColumnName(col.name),
          value: col.name,
          sortable: true,
          align: 'start',
        })
      }

      headers.push({ text: '', value: 'actions', sortable: false, align: 'end' })

      return headers
    })

    const columnMap = computed((): Record<string, MetricColumn> => {
      const colMap: Record<string, MetricColumn> = {}
      updateColumnMap(colMap, props.columns)
      return colMap
    })

    const internalAxiosParams = computed(() => {
      if (!props.isResolved) {
        return { _: undefined }
      }
      return props.axiosParams
    })

    function hasColumn(name: string): boolean {
      if (props.groups.length) {
        const item = props.groups[0]
        return name in item
      }

      const index = props.columns.findIndex((col) => col.name === name)
      return index >= 0
    }

    function shortColumnName(name: string) {
      switch (name) {
        case AttrKey.spanErrorCount:
          return 'errors'
        case AttrKey.spanErrorRate:
          return 'err%'
      }

      let m = name.match(/^([0-9a-z]+)\(\.duration\)$/)
      if (m) {
        return m[1]
      }

      const spanPrefix = 'span.'
      if (name.startsWith(spanPrefix)) {
        name = name.slice(spanPrefix.length)
      }

      m = name.match(/^(\w+)_per_(\w+)$/)
      if (m) {
        return `${m[1]}/${m[2]}`
      }

      return name
    }

    function eventsModeFor(group: Group) {
      const system = group[AttrKey.spanSystem]
      if (system) {
        return isEventSystem(system)
      }
      return props.eventsMode
    }

    return {
      AttrKey,

      groupingColumns,
      plainColumns,
      headers,
      columnMap,
      internalAxiosParams,

      eventsModeFor,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-data-table ::v-deep .v-data-table__wrapper > table {
  & > tbody > tr > td,
  & > tbody > tr > th,
  & > thead > tr > td,
  & > thead > tr > th,
  & > tfoot > tr > td,
  & > tfoot > tr > th {
    padding: 0 8px;
  }
}
</style>
