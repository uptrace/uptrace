<template>
  <v-simple-table :class="{ 'v-data-table--narrow': columns.length > 5 }">
    <thead v-if="items.length" class="v-data-table-header">
      <tr>
        <ThOrder v-for="colName in grouping" :key="colName" :value="colName" :order="order">
          {{ colName }}
        </ThOrder>
        <ThOrder v-for="col in aggColumns" :key="col.nme" :value="col.name" :order="order">
          {{ col.name }}
        </ThOrder>
      </tr>
    </thead>

    <thead v-if="loading">
      <tr class="v-data-table__progress">
        <th colspan="99" class="column">
          <v-progress-linear height="2" absolute indeterminate />
        </th>
      </tr>
    </thead>

    <tbody v-if="!items.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99">There are no matching timeseries...</td>
      </tr>
    </tbody>

    <tbody>
      <MetricItemsTableRow
        v-for="(item, i) in items"
        :key="i"
        :axios-params="axiosParams"
        :query="item[AttrKey.itemQuery]"
        :column-map="columnMap"
        :class="{ 'cursor-pointer': 'click' in $listeners }"
        @click="$emit('click', item)"
      >
        <template #default="{ metrics, defaultTimeseries }">
          <td v-for="colName in grouping" :key="colName">{{ item[colName] }}</td>
          <td v-for="col in aggColumns" :key="col.name" class="text-subtitle-2">
            <div class="d-flex align-center">
              <SparklineChart
                :name="col.name"
                :unit="col.unit"
                :line="(metrics[col.name] && metrics[col.name].value) || defaultTimeseries.value"
                :time="(metrics[col.name] && metrics[col.name].time) || defaultTimeseries.time"
                class="mr-2"
              />
              <XNum :value="item[col.name] || 0" :unit="col.unit" title="{0} per minute" />
            </div>
          </td>
        </template>
      </MetricItemsTableRow>
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'
import { TableItem } from '@/metrics/use-query'
import { MetricColumn, ColumnInfo } from '@/metrics/types'

// Components
import ThOrder from '@/components/ThOrder.vue'
import SparklineChart from '@/components/SparklineChart.vue'
import MetricItemsTableRow from '@/metrics/MetricItemsTableRow.vue'

// Utilities
import { AttrKey } from '@/models/otelattr'

export default defineComponent({
  name: 'MetricItemsTable',
  components: { ThOrder, SparklineChart, MetricItemsTableRow },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array as PropType<TableItem[]>,
      required: true,
    },
    columns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      required: true,
    },
  },

  setup(props) {
    const aggColumns = computed(() => {
      return props.columns
        .filter((col) => !col.isGroup)
        .map((col) => {
          return {
            ...col,
            unit: props.columnMap[col.name]?.unit ?? col.unit,
          }
        })
    })

    const grouping = computed((): string[] => {
      return props.columns.filter((col) => col.isGroup).map((col) => col.name)
    })

    return { AttrKey, aggColumns, grouping }
  },
})
</script>

<style lang="scss" scoped></style>
