<template>
  <v-expansion-panels v-model="panels" :flat="flat" multiple>
    <v-expansion-panel v-for="(event, i) in events" :key="i">
      <v-expansion-panel-header class="user-select-text">
        <span>
          <XDate :date="event.time" format="time" class="mr-5 text-caption" />
          <span>{{ event.eventName }}</span>
        </span>
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <EventPanelContent :date-range="dateRange" :event="event" />
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script lang="ts">
import { defineComponent, ref, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import EventPanelContent from '@/components/EventPanelContent.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { Span } from '@/models/span'

export default defineComponent({
  name: 'EventPanels',
  components: { EventPanelContent },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    events: {
      type: Array as PropType<Span[]>,
      required: true,
    },
    flat: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const panels = ref<number[]>([])

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

    return { xkey, panels }
  },
})
</script>

<style lang="scss" scoped></style>
