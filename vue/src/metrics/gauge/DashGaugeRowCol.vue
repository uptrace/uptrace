<template>
  <v-col cols="auto">
    <DashGaugeCard
      :loading="gaugeQuery.loading"
      :dash-gauge="internalDashGauge"
      :columns="gaugeQuery.columns"
      :values="gaugeQuery.values"
      :column-map="dashGauge.columnMap"
      show-edit
      :editable="editable"
      @click:edit="$emit('click:edit', $event)"
      @change="$emit('change', $event)"
    ></DashGaugeCard>
  </v-col>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useDashGaugeQuery } from '@/metrics/gauge/use-dash-gauges'

// Components
import DashGaugeCard from '@/metrics/gauge/DashGaugeCard.vue'

// Utilities
import { DashGauge } from '@/metrics/types'

export default defineComponent({
  name: 'DashGaugeRowCol',
  components: { DashGaugeCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashGauge: {
      type: Object as PropType<DashGauge>,
      required: true,
    },
    gridQuery: {
      type: String,
      default: '',
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const internalDashGauge = computed(() => {
      const dashGauge = cloneDeep(props.dashGauge)
      if (props.gridQuery) {
        dashGauge.query += ` | ${props.gridQuery}`
      }
      return dashGauge
    })

    const gaugeQuery = useDashGaugeQuery(
      () => {
        if (!internalDashGauge.value.metrics.length) {
          return { _: undefined }
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: internalDashGauge.value.metrics.map((m) => m.name),
          alias: internalDashGauge.value.metrics.map((m) => m.alias),
          query: internalDashGauge.value.query,
        }
      },
      computed(() => props.dashGauge.columnMap),
    )

    return { internalDashGauge, gaugeQuery }
  },
})
</script>

<style lang="scss" scoped></style>
