<template>
  <v-data-table
    :loading="loading"
    :headers="headers"
    :items="items"
    :items-per-page="itemsPerPage"
    :hide-default-footer="items.length <= 5"
    :sort-by.sync="order.column"
    :sort-desc.sync="order.desc"
    must-sort
    no-data-text="There are no metrics"
    class="v-data-table--narrow"
  >
    <template #item="{ item }">
      <TimeseriesTableRow
        :axios-params="axiosParams"
        :query="item._query"
        :class="{ 'cursor-pointer': 'click' in $listeners }"
        @click="$emit('click', item)"
      >
        <template #default="{ metrics, value, time }">
          <template v-for="attrKey in grouping">
            <td v-if="attrKey === AttrKey.spanGroupId" :key="attrKey">
              <router-link :to="spanListRouteFor(item)">{{ eventOrSpanName(item) }}</router-link>
            </td>
            <td v-else :key="attrKey">{{ item[attrKey] }}</td>
          </template>

          <td v-for="col in aggColumns" :key="col.name" class="text-subtitle-2">
            <div class="d-flex align-center">
              <SparklineChart
                :name="col.name"
                :line="(metrics[col.name] && metrics[col.name].value) || value"
                :time="time"
                :unit="col.unit"
                :color="col.color"
                :group="item._query"
                class="mr-2"
              />
              <XNum :value="item[col.name] || 0" :unit="col.unit" />
            </div>
          </td>
        </template>
      </TimeseriesTableRow>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'
import { exploreAttr } from '@/use/uql'
import { AxiosParams } from '@/use/watch-axios'
import { TableItem } from '@/metrics/use-query'

// Components
import SparklineChart from '@/components/SparklineChart.vue'
import TimeseriesTableRow from '@/metrics/TimeseriesTableRow.vue'

// Utilities
import { StyledColumnInfo } from '@/metrics/types'
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'TimeseriesTable',
  components: { SparklineChart, TimeseriesTableRow },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array as PropType<TableItem[]>,
      required: true,
    },
    itemsPerPage: {
      type: Number,
      default: 10,
    },
    columns: {
      type: Array as PropType<StyledColumnInfo[]>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      default: undefined,
    },
  },

  setup(props) {
    const grouping = computed((): string[] => {
      return props.columns.filter((col) => col.isGroup).map((col) => col.name)
    })

    const aggColumns = computed(() => {
      return props.columns.filter((col) => !col.isGroup)
    })

    const headers = computed(() => {
      const headers = []
      for (let colName of grouping.value) {
        headers.push({ text: colName, value: colName, sortable: true })
      }
      for (let col of aggColumns.value) {
        headers.push({ text: col.name, value: col.name, sortable: true, align: 'start' })
      }
      return headers
    })

    function spanListRouteFor(item: TableItem) {
      const query = exploreAttr(AttrKey.spanGroupId)
      const groupId = item[AttrKey.spanGroupId]
      return {
        name: 'SpanList',
        query: {
          query: `${query} | where ${AttrKey.spanGroupId} = ${groupId}`,
        },
      }
    }

    return { AttrKey, grouping, aggColumns, headers, spanListRouteFor, eventOrSpanName }
  },
})

function eventOrSpanName(item: Record<string, any>, maxLength = 120) {
  const eventName = item[AttrKey.spanEventName]
  if (eventName) {
    return truncate(eventName, { length: maxLength })
  }
  return truncate(item[AttrKey.spanName], { length: maxLength })
}
</script>

<style lang="scss" scoped></style>
