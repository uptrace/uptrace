<template>
  <UptraceQuery
    :uql="uql"
    hint="Select some metrics and use Aggregate button to plot something..."
    :disabled="disabled"
  >
    <div class="d-flex align-center">
      <MetricsAggMenu v-if="showAgg" :metrics="metrics" :uql="uql" :disabled="disabled" />
      <DashGroupingMenu
        v-if="showGroupBy"
        :uql="uql"
        :attr-keys="keysDs.values"
        :disabled="disabled"
      />
      <DashWhereBtn v-if="showDashWhere" :uql="uql" :axios-params="axiosParams" />

      <QueryHelpDialog />
      <v-btn text class="v-btn--filter" @click="uql.rawMode = !uql.rawMode">{{
        uql.rawMode ? 'Cancel' : 'Edit'
      }}</v-btn>
    </div>
  </UptraceQuery>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { ActiveMetric } from '@/metrics/types'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import DashGroupingMenu from '@/metrics/query/DashGroupingMenu.vue'
import DashWhereBtn from '@/metrics/query/DashWhereBtn.vue'
import MetricsAggMenu from '@/metrics/query/MetricsAggMenu.vue'
import QueryHelpDialog from '@/metrics/query/QueryHelpDialog.vue'

export default defineComponent({
  name: 'MetricsQueryBuilder',
  components: {
    UptraceQuery,
    DashGroupingMenu,
    MetricsAggMenu,
    DashWhereBtn,
    QueryHelpDialog,
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
    showAgg: {
      type: Boolean,
      default: false,
    },
    showGroupBy: {
      type: Boolean,
      default: false,
    },
    showDashWhere: {
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
        metric: props.metrics.map((metric) => metric.name),
      }
    })

    const keysDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attributes`,
        params: axiosParams.value,
      }
    })

    return { axiosParams, keysDs }
  },
})
</script>

<style lang="scss" scoped></style>
