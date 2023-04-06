<template>
  <v-expansion-panels v-model="panels" :flat="flat" multiple>
    <v-expansion-panel v-for="(event, i) in events" :key="i" :readonly="!hasAttrs(event)">
      <v-expansion-panel-header class="user-select-text">
        <span>
          <XDate :date="event.time" format="time" class="mr-5 text-caption" />
          <span>{{ event.name }}</span>
        </span>
      </v-expansion-panel-header>
      <v-expansion-panel-content v-if="hasAttrs(event)">
        <EventPanelContent :date-range="dateRange" :event="event" />
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import EventPanelContent from '@/tracing/EventPanelContent.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { SpanEvent } from '@/models/span'

export default defineComponent({
  name: 'EventPanels',
  components: { EventPanelContent },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    events: {
      type: Array as PropType<SpanEvent[]>,
      required: true,
    },
    flat: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const panels = shallowRef<number[]>([])

    watch(
      () => props.events,
      (events) => {
        if (events.length === 1) {
          panels.value = [0]
        } else {
          panels.value = []
        }
      },
      { immediate: true },
    )

    function hasAttrs(event: SpanEvent): boolean {
      return Boolean(event.attrs && Object.keys(event.attrs).length)
    }

    return { AttrKey, panels, hasAttrs }
  },
})
</script>

<style lang="scss" scoped></style>
