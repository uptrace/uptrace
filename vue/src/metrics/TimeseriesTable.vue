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
    :footer-props="{ disableItemsPerPage: true }"
    :dense="dense"
    class="v-data-table--narrow"
    @current-items="
      currentItems = $event
      $emit('current-items', $event)
    "
    @update:sort-by="$nextTick(() => (order.desc = true))"
  >
    <template #no-data>
      <div class="pa-4">
        <div class="mb-4">The query result is empty.</div>
      </div>
    </template>
    <template #item="{ item }">
      <TimeseriesTableRow
        :axios-params="axiosParams"
        :query="item._query"
        :class="{ 'cursor-pointer': 'click' in $listeners }"
        @click="$emit('click', item)"
      >
        <template #default="{ metrics, time, emptyValue }">
          <template v-for="attrKey in groupingColumns">
            <td v-if="attrKey === AttrKey.spanGroupId" :key="attrKey">
              <router-link :to="routeForSpanList(item[AttrKey.spanGroupdId])">{{
                item[AttrKey.displayName] || item[AttrKey.spanGroupId]
              }}</router-link>
            </td>
            <td v-else :key="attrKey">{{ item[attrKey] }}</td>
          </template>

          <td v-for="col in aggColumns" :key="col.name" class="text-subtitle-2">
            <div class="d-flex align-center">
              <SparklineChart
                :name="col.name"
                :line="(metrics[col.name] && metrics[col.name].value) || emptyValue"
                :time="time"
                :unit="col.unit"
                :color="col.color"
                class="mr-2"
              />
              <NumValue :value="item[col.name] || 0" :unit="col.unit" />
            </div>
          </td>
        </template>
      </TimeseriesTableRow>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'
import { exploreAttr } from '@/use/uql'

// Components
import SparklineChart from '@/components/SparklineChart.vue'
import TimeseriesTableRow from '@/metrics/TimeseriesTableRow.vue'

// Misc
import { StyledColumnInfo, TableRowData } from '@/metrics/types'
import { AttrKey } from '@/models/otel'
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'TimeseriesTable',
  components: { SparklineChart, TimeseriesTableRow },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array as PropType<TableRowData[]>,
      required: true,
    },
    itemsPerPage: {
      type: Number,
      default: 15,
    },
    aggColumns: {
      type: Array as PropType<StyledColumnInfo[]>,
      required: true,
    },
    groupingColumns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },
    dense: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const headers = computed(() => {
      const headers = []
      for (let colName of props.groupingColumns) {
        headers.push({ text: colName, value: colName, sortable: true })
      }
      for (let col of props.aggColumns) {
        headers.push({ text: col.name, value: col.name, sortable: true, align: 'start' })
      }
      return headers
    })

    function routeForSpanList(groupId: string) {
      const query = exploreAttr(AttrKey.spanGroupId, true)
      return {
        name: 'SpanList',
        query: {
          query: `${query} | where ${AttrKey.spanGroupId} = ${groupId}`,
        },
      }
    }

    return {
      Unit,
      AttrKey,

      headers,

      routeForSpanList,
    }
  },
})
</script>

<style lang="scss" scoped></style>
