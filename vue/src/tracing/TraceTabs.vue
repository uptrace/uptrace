<template>
  <v-card v-if="trace" outlined rounded="lg">
    <v-tabs v-model="activeTab" background-color="transparent" class="light-blue lighten-5">
      <v-tab href="#trace">Trace</v-tab>
      <v-tab v-for="(events, system) in trace.events" :key="system" :href="`#${system}`">
        {{ system }} ({{ events.length }})
      </v-tab>
    </v-tabs>

    <v-tabs-items v-model="activeTab">
      <v-tab-item value="trace" class="px-4 py-6">
        <v-container fluid>
          <v-row align="center">
            <v-col v-if="rootSpanId" cols="auto">
              <v-chip close color="primary" outlined @click:close="$emit('click:crop', undefined)">
                <span>Focused on span: </span>
                <strong class="ml-1">{{ rootSpanId }}</strong>
              </v-chip>
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <TraceTimeline
                :date-range="dateRange"
                :trace="trace"
                :root-span-id="rootSpanId"
                @click:crop="$emit('click:crop', $event)"
              />
            </v-col>
          </v-row>
        </v-container>
      </v-tab-item>
      <v-tab-item v-for="(events, system) in trace.events" :key="system" :value="system">
        <EventPanels :date-range="dateRange" :events="events" :annotations="annotations" />
      </v-tab-item>
    </v-tabs-items>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from 'vue'

// Composables
import { UseTrace } from '@/tracing/use-trace'
import { UseDateRange } from '@/use/date-range'
import { Annotation } from '@/org/use-annotations'

// Components
import TraceTimeline from '@/tracing/TraceTimeline.vue'
import EventPanels from '@/tracing/EventPanels.vue'

export default defineComponent({
  name: 'TraceTabs',
  components: {
    TraceTimeline,
    EventPanels,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    trace: {
      type: Object as PropType<UseTrace>,
      default: undefined,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
    rootSpanId: {
      type: String,
      default: undefined,
    },
  },

  setup() {
    const activeTab = ref()

    return {
      activeTab,
    }
  },
})
</script>

<style lang="scss" scoped></style>
