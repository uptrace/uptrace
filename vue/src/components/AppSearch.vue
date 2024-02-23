<template>
  <div>
    <v-text-field
      v-if="value"
      ref="textField"
      v-model="searchInput"
      prepend-inner-icon="mdi-magnify"
      placeholder="Search or jump to trace id..."
      hide-details
      flat
      solo
      background-color="grey lighten-4"
      style="width: 360px"
      @keyup.enter="submit"
      @keyup.esc="hideSearch"
      @blur="hideSearch"
    />

    <v-btn v-else icon @click="showSearch">
      <v-icon>mdi-magnify</v-icon>
    </v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, nextTick } from 'vue'

// Composables
import { useRouterOnly } from '@/use/router'

const TRACE_ID_RE = /^([0-9a-f]{32}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/i

export default defineComponent({
  name: 'AppSearch',

  props: {
    value: {
      type: Boolean,
      required: true,
    },
  },

  setup(props, ctx) {
    const router = useRouterOnly()

    const textField = shallowRef()
    const searchInput = shallowRef('')

    function showSearch() {
      ctx.emit('input', true)
      nextTick(() => {
        textField.value.focus()
      })
    }

    function hideSearch() {
      ctx.emit('input', false)
    }

    function submit() {
      if (!searchInput.value) {
        return
      }

      const str = searchInput.value.trim()
      searchInput.value = ''
      hideSearch()

      if (TRACE_ID_RE.test(str)) {
        router.push({
          name: 'TraceFind',
          params: { traceId: str },
        })
        return
      }

      router
        .push({
          name: 'SpanGroupList',
          query: {
            search: str,
          },
        })
        .catch(() => {})
    }

    return {
      textField,
      searchInput,

      showSearch,
      hideSearch,

      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
