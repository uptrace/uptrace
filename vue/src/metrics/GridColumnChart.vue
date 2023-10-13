<template>
  <div>
    <v-row no-gutters justify="space-around">
      <v-col :cols="legend.placement === LegendPlacement.Bottom ? 12 : ''">
        <MetricChart
          :annotations="annotations"
          :loading="loading"
          :resolved="resolved"
          :timeseries="activeTimeseries"
          :time="time"
          :chart-kind="chartKind"
          :height="chartHeight"
          :event-bus="eventBus"
        />
      </v-col>
      <v-col
        v-if="legend.type !== LegendType.None"
        v-element-resize
        cols="auto"
        class="pr-2"
        @resize="onLegendResize"
      >
        <ChartLegendTable
          v-if="legend.type === LegendType.Table"
          :loading="loading"
          :timeseries="timeseries"
          :values="legend.values"
          :max-length="legend.maxLength ?? 40"
          @current-items="currentTimeseries = $event"
          @hover:item="eventBus.emit('hover', $event)"
        />
        <ChartLegendList
          v-else-if="legend.type === LegendType.List"
          :loading="loading"
          :timeseries="timeseries"
          :values="legend.values"
          :direction="legend.placement === LegendPlacement.Bottom ? 'row' : 'column'"
          @current-items="currentTimeseries = $event"
          @hover:item="eventBus.emit('hover', $event)"
        />
      </v-col>
    </v-row>

    <v-row v-if="showDuplicateLegend" dense justify="space-around">
      <v-col cols="auto">
        <ChartLegendTable :loading="timeseries.loading" :timeseries="timeseries">
          <template #expanded-item="{ headers, item, expandItem }">
            <slot name="expanded-item" :headers="headers" :item="item" :expand-item="expandItem" />
          </template>
        </ChartLegendTable>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { injectAnnotations } from '@/org/use-annotations'

// Components
import MetricChart from '@/metrics/MetricChart.vue'
import ChartLegendTable from '@/metrics/ChartLegendTable.vue'
import ChartLegendList from '@/metrics/ChartLegendList.vue'

// Utilities
import { EventBus } from '@/models/eventbus'
import {
  ChartKind,
  ChartLegend,
  LegendType,
  LegendPlacement,
  StyledTimeseries,
} from '@/metrics/types'

export default defineComponent({
  name: 'GridColumnChart',
  components: {
    ChartLegendTable,
    ChartLegendList,
    MetricChart,
  },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    resolved: {
      type: Boolean,
      required: true,
    },
    timeseries: {
      type: Array as PropType<StyledTimeseries[]>,
      required: true,
    },
    time: {
      type: Array as PropType<string[]>,
      required: true,
    },
    chartKind: {
      type: String as PropType<ChartKind>,
      default: ChartKind.Line,
    },
    legend: {
      type: Object as PropType<ChartLegend>,
      required: true,
    },
    height: {
      type: Number,
      default: 200,
    },
  },

  setup(props, ctx) {
    const eventBus = new EventBus()

    const legendHeight = shallowRef(0)
    function onLegendResize(event: any) {
      legendHeight.value = event.detail.height
    }

    const showDuplicateLegend = computed(() => {
      if (
        props.legend.placement === LegendPlacement.Bottom &&
        props.legend.type === LegendType.Table
      ) {
        return false
      }
      return 'expanded-item' in ctx.slots
    })

    const chartHeight = computed(() => {
      const minHeight = 140

      if (
        props.legend.type === LegendType.None ||
        props.legend.placement === LegendPlacement.Right
      ) {
        return props.height
      }

      switch (props.legend.type) {
        case LegendType.Table: {
          const height = props.height - legendHeight.value
          return Math.max(height, minHeight)
        }
        case LegendType.List: {
          const height = props.height - legendHeight.value
          return Math.max(height, minHeight)
        }
        default:
          console.error(`unsupported legend type: ${props.legend.type}`)
          return props.height
      }
    })

    const currentTimeseries = shallowRef<StyledTimeseries[]>()
    const activeTimeseries = computed(() => {
      if (props.legend.type !== LegendType.None && currentTimeseries.value) {
        return currentTimeseries.value
      }
      return props.timeseries
    })

    return {
      LegendType,
      LegendPlacement,

      annotations: injectAnnotations(),

      showDuplicateLegend,
      chartHeight,

      eventBus,
      currentTimeseries,
      activeTimeseries,

      onLegendResize,
    }
  },
})
</script>

<style lang="scss" scoped></style>
