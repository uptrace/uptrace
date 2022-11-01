<template>
  <v-text-field
    v-model="traceId"
    prepend-inner-icon="mdi-magnify"
    placeholder="Search or jump to trace id..."
    hide-details
    flat
    solo
    background-color="grey lighten-4"
    style="min-width: 300px; width: 300px"
    @keyup.enter="jumpToTrace"
  />
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { createUqlEditor } from '@/use/uql'

// Utilities
import { AttrKey } from '@/models/otel'

const TRACE_ID_RE = /^[0-9A-Fa-f]{32}$/

export default defineComponent({
  name: 'Search',

  setup() {
    const { router } = useRouter()
    const traceId = shallowRef('')

    function jumpToTrace() {
      traceId.value = traceId.value.trim()
      if (TRACE_ID_RE.test(traceId.value)) {
        router.push({
          name: 'TraceFind',
          params: { traceId: traceId.value },
        })
      } else {
        const query = createUqlEditor()
          .exploreAttr(AttrKey.spanGroupId)
          .where(`{${AttrKey.spanName},${AttrKey.spanEventName}}`, 'contains', traceId.value)
          .toString()
        router.push({
          name: 'SpanGroupList',
          params: { traceId: traceId.value },
          query: {
            query,
          },
        })
      }
      traceId.value = ''
    }

    return { traceId, jumpToTrace }
  },
})
</script>

<style lang="scss" scoped></style>
