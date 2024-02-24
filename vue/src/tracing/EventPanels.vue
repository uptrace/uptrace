<template>
  <v-expansion-panels v-model="panels" :flat="flat" multiple>
    <v-expansion-panel v-for="(event, i) in events" :key="i" :readonly="!hasAttrs(event)">
      <v-expansion-panel-header class="user-select-text">
        <span>
          <DateValue :value="event.time" format="time" class="mr-5 text-caption" />

          <v-btn v-if="event.span" icon @click.stop="$emit('click:span', event.span)">
            <v-icon>mdi-link</v-icon>
          </v-btn>

          <span class="text-subtitle-1">{{ event.name }}</span>
          <template v-if="event.span">
            <span class="mx-2"> &bull; </span>
            <span class="text-subtitle-1">{{ event.span.displayName }}</span>
          </template>
        </span>
      </v-expansion-panel-header>
      <v-expansion-panel-content v-if="hasAttrs(event)">
        <EventPanelContent :date-range="dateRange" :event="event" :annotations="annotations">
          <template #append-action>
            <v-btn
              v-if="event.span"
              depressed
              small
              class="ml-2"
              @click="$emit('click:span', event.span)"
            >
              Go to span
            </v-btn>
          </template>
        </EventPanelContent>
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

// Misc
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
    annotations: {
      type: Array,
      default: () => [],
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

    return { panels, hasAttrs }
  },
})
</script>

<style lang="scss" scoped></style>
