<template>
  <div>
    <portal to="navigation">
      <SystemGroupPicker :loading="systems.loading" :systems="systems.items" optional />
    </portal>

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
import { injectForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useSystems } from '@/tracing/system/use-systems'
import { useSpan } from '@/tracing/use-spans'

// Components
import SystemGroupPicker from '@/tracing/system/SystemGroupPicker.vue'
import SpanCard from '@/tracing/SpanCard.vue'

export default defineComponent({
  name: 'TraceSpanShow',
  components: { SystemGroupPicker, SpanCard },

  setup() {
    const route = useRoute()
    const dateRange = useDateRange()
    const systems = useSystems(() => {
      return {
        ...dateRange.axiosParams(),
      }
    })

    const forceReload = injectForceReload()
    const span = useSpan(() => {
      const { projectId, traceId, spanId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/traces/${traceId}/${spanId}`,
        params: forceReload.params,
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

    return { dateRange, systems, span }
  },
})
</script>

<style lang="scss" scoped></style>
