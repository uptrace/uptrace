<template>
  <v-data-table
    :headers="headers"
    :items="groups"
    :items-per-page="9"
    :hide-default-footer="groups.length <= 9"
    no-data-text="There are no metrics"
    item-key="_id"
    show-select
    :sort-by.sync="order.column"
    :sort-desc.sync="order.desc"
    must-sort
    show-expand
    single-expand
    class="v-data-table--narrow"
    @current-items="$emit('current-items', $event)"
  >
    <template #[`header.data-table-select`]>
      <v-simple-checkbox
        :value="hasSelected"
        :ripple="false"
        class="mr-0"
        @click="toggleSelected"
      ></v-simple-checkbox>
    </template>
    <template #item="{ item, isExpanded, expand }">
      <tr
        class="cursor-pointer"
        @mouseenter="$emit('hover:item', { item: { id: item._id }, hover: true })"
        @mouseleave="$emit('hover:item', { item: { id: item._id }, hover: false })"
        @click="expand(!isExpanded)"
      >
        <td>
          <v-simple-checkbox
            :value="item._selected"
            :color="item._color"
            :ripple="false"
            @click.stop="item._selected = !item._selected"
          ></v-simple-checkbox>
        </td>
        <td v-for="col in groupingColumns" :key="col.name">
          <span v-if="col.name === AttrKey.spanGroupId && item._name">
            {{ truncate(item._name, { length: 120 }) }}
          </span>
          <span v-else>
            <AnyValue :value="item[col.name]" :name="col.name" />
          </span>
        </td>
        <td v-for="col in metricColumns" :key="col.name" class="text-right">
          <XNum :value="item['_avg_' + col.name]" :unit="col.unit" />
        </td>
        <td>
          <v-btn v-if="isExpanded" icon title="Hide spans" @click.stop="expand(false)">
            <v-icon>mdi-chevron-up</v-icon>
          </v-btn>
          <v-btn v-else icon title="View spans" @click.stop="expand(true)">
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </td>
      </tr>
    </template>
    <template #expanded-item="{ headers, item }">
      <tr class="v-data-table__expanded v-data-table__expanded__content">
        <td :colspan="headers.length" class="pa-4">
          <SpansList :events-mode="eventsMode" :axios-params="axiosParams" :where="item._query" />
        </td>
      </tr>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'
import { UseUql } from '@/use/uql'
import { TimeseriesGroup, ColumnInfo } from '@/tracing/use-timeseries'

// Components
import SpansList from '@/tracing/SpansList.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'TimeseriesGroupsTable',
  components: { SpansList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      default: undefined,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },

    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    groups: {
      type: Array as PropType<TimeseriesGroup[]>,
      required: true,
    },
    groupingColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    metricColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const headers = computed(() => {
      const headers = []
      for (let col of props.groupingColumns) {
        headers.push({ text: col.name, value: col.name, sortable: true, align: 'start' })
      }
      for (let col of props.metricColumns) {
        headers.push({ text: col.name, value: '_avg_' + col.name, sortable: true, align: 'end' })
      }
      headers.push({ text: '', value: 'data-table-expand', sortable: false })
      return headers
    })

    const hasSelected = computed((): boolean => {
      return props.groups.some((item) => item._selected)
    })

    function toggleSelected() {
      const selected = !hasSelected.value
      props.groups.forEach((item) => {
        item._selected = selected
      })
    }

    return {
      AttrKey,

      headers,

      hasSelected,
      toggleSelected,
      truncate,
    }
  },
})
</script>

<style lang="scss" scoped></style>
