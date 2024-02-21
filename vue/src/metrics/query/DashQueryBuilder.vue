<template>
  <UptraceQuery :uql="uql">
    <DashWhereBtn :uql="uql" :axios-params="axiosParams" />

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
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import DashWhereBtn from '@/metrics/query/DashWhereBtn.vue'

export default defineComponent({
  name: 'DashQueryBuilder',
  components: { UptraceQuery, DashWhereBtn },

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
  },

  setup(props) {
    const axiosParams = computed(() => {
      if (!props.metrics.length) {
        return { _: undefined }
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.metrics,
      }
    })

    return {
      axiosParams,
    }
  },
})
</script>

<style lang="scss" scoped></style>
