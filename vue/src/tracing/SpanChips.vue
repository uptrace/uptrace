<template>
  <span>
    <v-btn
      v-if="span.statusCode === 'error'"
      icon
      :title="`${AttrKey.spanStatusCode} = 'error'`"
      class="mr-1"
      :class="{ 'cursor-default': !clickable }"
      @click.stop="$emit('click:chip', { key: AttrKey.spanStatusCode, value: 'error' })"
    >
      <v-icon color="error"> mdi-alert-circle-outline </v-icon>
    </v-btn>

    <v-chip
      v-for="(chip, i) in chips"
      :key="chip.key"
      color="light-blue lighten-5"
      label
      small
      :title="`${chip.key}: ${chip.value}`"
      class="mb-1"
      :class="{ 'ml-1': i > 0, 'cursor-default': !clickable }"
      @click.stop="$emit('click:chip', chip)"
    >
      {{ chip.text }}
    </v-chip>

    <v-chip v-if="events.length" color="blue lighten-5" label small class="ml-1">
      <strong class="mr-1">{{ events.length }}</strong>
      <span>{{ events.length === 1 ? 'event' : 'events' }}</span>
    </v-chip>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { AttrMap, Span } from '@/models/span'

export interface SpanChip {
  key: string
  value: any
  text: string
}

export default defineComponent({
  name: 'SpanChips',

  props: {
    span: {
      type: Object as PropType<Span>,
      required: true,
    },
    showOperation: {
      type: Boolean,
      default: false,
    },
    traceMode: {
      type: Boolean,
      default: false,
    },
    clickable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const events = computed((): Span[] => {
      return props.span?.events ?? []
    })

    const chips = computed(() => {
      const chips: SpanChip[] = []

      const service = props.span.attrs[AttrKey.serviceName]
      if (service) {
        chips.push({ key: AttrKey.serviceName, value: service, text: service })
      }

      if (props.traceMode) {
        const spanSystem = props.span.system
        if (!spanSystem.endsWith(`:${service}`)) {
          chips.push({ key: AttrKey.spanSystem, value: spanSystem, text: spanSystem })
        }
      }

      pushHttpStatusChip(chips, props.span.attrs)

      return chips
    })

    return { AttrKey, events, chips }
  },
})

function pushHttpStatusChip(chips: SpanChip[], attrs: AttrMap) {
  const httpCode = attrs[AttrKey.httpStatusCode]
  if (typeof httpCode === 'number' && httpCode != 0 && (httpCode < 200 || httpCode >= 300)) {
    chips.push({ key: AttrKey.httpStatusCode, value: httpCode, text: String(httpCode) })
  }
}
</script>

<style lang="scss" scoped>
.cursor-default {
  cursor: default;
}
</style>
