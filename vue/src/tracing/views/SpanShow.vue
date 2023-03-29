<template>
  <div>
    <v-progress-linear v-if="span.loading" absolute indeterminate />
    <SpanCard
      v-if="span.data"
      :date-range="dateRange"
      :span="span.data"
      :fluid="$vuetify.breakpoint.mdAndDown"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, watch } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { useForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useSpan } from '@/tracing/use-spans'

// Components
import SpanCard from '@/tracing/SpanCard.vue'

export default defineComponent({
  name: 'SpanShow',
  components: { SpanCard },

  setup() {
    const route = useRoute()
    const dateRange = useDateRange()
    const { forceReloadParams } = useForceReload()

    const span = useSpan(() => {
      const { projectId, traceId, spanId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/traces/${traceId}/${spanId}`,
        params: forceReloadParams.value,
      }
    })

    watch(
      () => span.data,
      (span) => {
        if (span) {
          useTitle(span.name)
        }
      },
    )

    return { dateRange, span }
  },
})
</script>

<style lang="scss" scoped></style>
