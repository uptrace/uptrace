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

    <v-chip v-if="traceId" :to="traceRoute" title="View trace" color="bg--none-primary" label small>
      {{ traceId }}
    </v-chip>

    <v-chip
      v-for="chip in chips"
      :key="chip.key"
      color="bg--none-primary"
      label
      small
      class="ml-1"
      :class="{ 'cursor-default': !clickable }"
      :title="`${chip.key}: ${chip.value}`"
      @click.stop="$emit('click:chip', chip)"
    >
      {{ chip.text }}
    </v-chip>

    <v-chip v-if="events.length" color="bg--none-primary" label small class="ml-1">
      <strong class="mr-1">{{ events.length }}</strong>
      <span>{{ events.length === 1 ? 'event' : 'events' }}</span>
    </v-chip>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Misc
import { isEventSystem, AttrKey } from '@/models/otel'
import { AttrMap, Span, SpanEvent } from '@/models/span'

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
    const traceId = computed(() => {
      if (props.traceMode) {
        return ''
      }
      if (props.span.standalone) {
        return ''
      }
      return props.span.traceId.slice(-6)
    })
    const traceRoute = computed(() => {
      if (!traceId.value) {
        return ''
      }
      return {
        name: 'TraceShow',
        params: {
          traceId: props.span.traceId,
        },
        query: {
          span: isEventSystem(props.span.system) ? props.span.parentId : props.span.id,
        },
      }
    })

    const events = computed((): SpanEvent[] => {
      return props.span?.events ?? []
    })

    const chips = computed(() => {
      const chips: SpanChip[] = []

      const env = props.span.attrs[AttrKey.deploymentEnvironment]
      if (env) {
        chips.push({ key: AttrKey.deploymentEnvironment, value: env, text: env })
      }

      const service = props.span.attrs[AttrKey.serviceName]
      if (service) {
        chips.push({ key: AttrKey.serviceName, value: service, text: service })
      }

      // Add kind to distinguish `client` and `server` spans.
      if (props.span.kind !== 'internal') {
        chips.push({ key: AttrKey.spanKind, value: props.span.kind, text: props.span.kind })
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

    return {
      AttrKey,

      traceId,
      traceRoute,

      events,
      chips,
    }
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
