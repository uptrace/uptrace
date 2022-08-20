<template>
  <UptraceQuery :uql="uql">
    <template v-if="metrics && dashAttrs.keys.length">
      <DashWhereMenu
        v-for="key in dashAttrs.keys.slice(0, 10)"
        :key="key"
        :uql="uql"
        :attr-key="key"
        :axios-params="axiosParams"
      />
    </template>
    <v-btn v-else text disabled class="v-btn--filter">Where</v-btn>

    <v-divider vertical class="mx-2" />
    <v-btn text class="v-btn--filter" @click="uql.rawMode = !uql.rawMode">{{
      uql.rawMode ? 'Cancel' : 'Edit'
    }}</v-btn>

    <template v-if="$slots['prepend-actions']" #prepend-actions>
      <slot name="prepend-actions" />
    </template>
  </UptraceQuery>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { useDashAttrs } from '@/metrics/use-dashboards'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import DashWhereMenu from '@/metrics/query/DashWhereMenu.vue'

export default defineComponent({
  name: 'DashQueryBuilder',
  components: { UptraceQuery, DashWhereMenu },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Array as PropType<string[]>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    showGroupBy: {
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
        metrics: props.metrics,
      }
    })

    const dashAttrs = useDashAttrs(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attributes`,
        params: axiosParams.value,
      }
    })

    return {
      axiosParams,
      dashAttrs,
    }
  },
})
</script>

<style lang="scss" scoped></style>
