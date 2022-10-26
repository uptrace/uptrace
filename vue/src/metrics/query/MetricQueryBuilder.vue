<template>
  <UptraceQuery
    :uql="uql"
    hint="Select some metrics and use Aggregate button to plot something..."
    :disabled="disabled"
  >
    <DashGroupByMenu
      v-if="showDashGroupBy"
      :uql="uql"
      :attr-keys="dashAttrs.keys"
      :disabled="disabled"
    />
    <MetricGroupByMenu
      v-if="showMetricGroupBy"
      :metrics="metrics"
      :uql="uql"
      :axios-params="axiosParams"
      :disabled="disabled"
    />
    <WhereMenu :metrics="metrics" :uql="uql" :axios-params="axiosParams" :disabled="disabled" />
    <AggMenu :metrics="metrics" :uql="uql" :disabled="disabled" />

    <v-divider vertical class="mx-2" />
    <v-btn text class="v-btn--filter" @click="uql.rawMode = !uql.rawMode">{{
      uql.rawMode ? 'Cancel' : 'Edit'
    }}</v-btn>
    <MetricQueryHelpDialog />
  </UptraceQuery>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { ActiveMetric } from '@/metrics/types'
import { useDashAttrs } from '@/metrics/use-dashboards'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import MetricGroupByMenu from '@/metrics/query/MetricGroupByMenu.vue'
import DashGroupByMenu from '@/metrics/query/DashGroupByMenu.vue'
import WhereMenu from '@/metrics/query/WhereMenu.vue'
import AggMenu from '@/metrics/query/AggMenu.vue'
import MetricQueryHelpDialog from '@/metrics/query/MetricQueryHelpDialog.vue'

export default defineComponent({
  name: 'MetricQueryBuilder',
  components: {
    UptraceQuery,
    MetricGroupByMenu,
    DashGroupByMenu,
    AggMenu,
    WhereMenu,
    MetricQueryHelpDialog,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Array as PropType<ActiveMetric[]>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    showDashGroupBy: {
      type: Boolean,
      default: false,
    },
    showMetricGroupBy: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const route = useRoute()

    const axiosParams = computed(() => {
      if (!props.metrics.length) {
        return { _: undefined }
      }

      return {
        ...props.dateRange.axiosParams(),
        metrics: props.metrics.map((metric) => metric.name),
      }
    })

    const dashAttrs = useDashAttrs(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attributes`,
        params: axiosParams.value,
      }
    })

    return { axiosParams, dashAttrs }
  },
})
</script>

<style lang="scss" scoped></style>
