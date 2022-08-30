<template>
  <SpanCard
    :date-range="dateRange"
    :span="span"
    :span-list-route="spanListRoute"
    :group-list-route="groupListRoute"
    fluid
    show-toolbar
  />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composable
import { useDateRange } from '@/use/date-range'

// Components
import SpanCard from '@/tracing/SpanCard.vue'

// Utilities
import { Span } from '@/models/span'

export default defineComponent({
  name: 'SpanCardDateRange',
  components: { SpanCard },

  props: {
    span: {
      type: Object as PropType<Span>,
      required: true,
    },
    spanListRoute: {
      type: String,
      required: true,
    },
    groupListRoute: {
      type: String,
      required: true,
    },
  },

  setup() {
    const dateRange = useDateRange()

    return { dateRange }
  },
})
</script>

<style lang="scss" scoped></style>
