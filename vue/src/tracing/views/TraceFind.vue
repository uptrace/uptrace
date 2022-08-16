<template>
  <v-container fluid class="fill-height">
    <v-card flat class="mx-auto">
      <v-card-text class="text-h5 text-center">
        Trace <strong>{{ $route.params.traceId }}</strong> not found.<br />
        It may take up to <strong>30 seconds</strong> to process a trace.
      </v-card-text>

      <v-card-actions class="justify-center">
        <v-btn :loading="traceSearch.loading" color="primary" @click="reload()">
          <v-icon left>mdi-refresh</v-icon>
          <span>Try again</span>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, watch } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useTraceSearch } from '@/tracing/use-trace-search'

export default defineComponent({
  name: 'TraceFind',

  setup() {
    const { router, route } = useRouter()
    const traceSearch = useTraceSearch()

    watch(
      () => traceSearch.trace,
      (trace) => {
        if (trace) {
          router.replace({
            name: 'TraceShow',
            params: {
              projectId: String(trace.projectId),
              traceId: trace.id,
            },
          })
        }
      },
    )

    watch(
      () => route.value.params.traceId,
      (traceId) => {
        traceSearch.find(traceId)
      },
      { immediate: true },
    )

    function reload() {
      traceSearch.find(route.value.params.traceId)
    }

    return { traceSearch, reload }
  },
})
</script>

<style lang="scss" scoped></style>
