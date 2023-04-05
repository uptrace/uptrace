<template>
  <div>
    <v-skeleton-loader v-if="!span.data" type="article, table" />

    <SpanBodyCard v-else :date-range="dateRange" :span="span.data" fluid>
      <template v-if="alert.createdAt !== alert.updatedAt" slot="append-column">
        <v-col cols="auto">
          <div class="grey--text font-weight-regular">First seen</div>
          <XDate :date="alert.createdAt" />
        </v-col>
      </template>
    </SpanBodyCard>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useSpan } from '@/tracing/use-spans'
import { ErrorAlert } from '@/alerting/use-alerts'

// Components
import SpanBodyCard from '@/tracing/SpanBodyCard.vue'

export default defineComponent({
  name: 'AlertCardSpan',
  components: {
    SpanBodyCard,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    alert: {
      type: Object as PropType<ErrorAlert>,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()

    const span = useSpan(() => {
      const { projectId } = route.value.params
      const { traceId, spanId } = props.alert.params
      return {
        url: `/api/v1/tracing/${projectId}/traces/${traceId}/${spanId}`,
        params: {
          ...props.dateRange.axiosParams(),
        },
      }
    })

    return { span }
  },
})
</script>

<style lang="scss" scoped></style>
