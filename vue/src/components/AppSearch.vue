<template>
  <v-text-field
    v-model="searchInput"
    prepend-inner-icon="mdi-magnify"
    placeholder="Search or jump to trace id..."
    hide-details
    flat
    solo
    background-color="grey lighten-4"
    style="min-width: 400px; width: 400px"
    @keyup.enter="submit"
  />
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { createQueryEditor } from '@/use/uql'

// Utilities
import { AttrKey } from '@/models/otel'

const TRACE_ID_RE = /^([0-9a-f]{32}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/i

export default defineComponent({
  name: 'Search',

  setup() {
    const { router } = useRouter()
    const searchInput = shallowRef('')

    function submit() {
      if (!searchInput.value) {
        return
      }

      const str = searchInput.value.trim()
      searchInput.value = ''

      if (TRACE_ID_RE.test(str)) {
        router.push({
          name: 'TraceFind',
          params: { traceId: str },
        })
        return
      }

      const query = createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId)
        .where(AttrKey.displayName, 'contains', str)
        .toString()
      router
        .push({
          name: 'SpanGroupList',
          query: {
            query,
          },
        })
        .catch(() => {})
    }

    return { searchInput, submit }
  },
})
</script>

<style lang="scss" scoped></style>
