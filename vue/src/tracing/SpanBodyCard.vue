<template>
  <div>
    <v-row class="px-2 text-subtitle-1">
      <v-col class="word-break-all">
        <template v-if="span.eventName">
          <span>{{ span.eventName }}</span>
          <template v-if="span.name">
            <span class="mx-2"> &bull; </span>
            <span>{{ spanName(span, 1000) }}</span>
          </template>
        </template>
        <span v-else>{{ spanName(span, 1000) }}</span>
      </v-col>
    </v-row>

    <v-row align="end" class="px-2 text-subtitle-2 text-center">
      <v-col v-if="span.attrs[AttrKey.deploymentEnvironment]" cols="auto">
        <div class="grey--text font-weight-regular">Env</div>
        <div>{{ span.attrs[AttrKey.deploymentEnvironment] }}</div>
      </v-col>

      <v-col v-if="span.attrs[AttrKey.serviceName]" cols="auto">
        <div class="grey--text font-weight-regular">Service</div>
        <div>{{ span.attrs[AttrKey.serviceName] }}</div>
      </v-col>

      <v-col v-if="span.kind" cols="auto">
        <div class="grey--text font-weight-regular">Kind</div>
        <div>{{ span.kind }}</div>
      </v-col>

      <v-col v-if="span.statusCode" cols="auto">
        <div class="grey--text font-weight-regular">Status</div>
        <div :class="{ 'error--text': span.statusCode === 'error' }">
          {{ span.statusCode }}
        </div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <XDate v-if="span.time" :date="span.time" format="full" />
      </v-col>

      <v-col v-if="span.duration > 0" cols="auto">
        <div class="grey--text font-weight-regular">Duration</div>
        <XDuration :duration="span.duration" fixed />
      </v-col>

      <v-col cols="auto">
        <div class="mb-0">
          <v-btn v-if="traceRoute" depressed small :to="traceRoute" exact> View trace </v-btn>
          <v-btn v-if="spanGroupRoute" depressed small :to="spanGroupRoute" exact class="ml-2">
            View spans
          </v-btn>

          <slot v-if="$slots['append-action']" name="append-action" />

          <NewMonitorMenu
            v-else
            :name="`${span.system} > ${eventOrSpanName(span)}`"
            :where="`where ${AttrKey.spanGroupId} = '${span.groupId}'`"
            :events-mode="isEvent"
            verbose
            class="ml-2"
          />
        </div>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-sheet outlined rounded="lg">
          <v-tabs v-model="activeTab" background-color="transparent" class="light-blue lighten-5">
            <v-tab href="#attrs">Attrs</v-tab>
            <v-tab v-if="dbStatement" href="#dbStatement">SQL:raw</v-tab>
            <v-tab v-if="dbStatementPretty" href="#dbStatementPretty">SQL:pretty</v-tab>
            <v-tab v-if="excStacktrace" href="#excStacktrace">Stacktrace</v-tab>
            <v-tab v-if="span.events && span.events.length" href="#events">
              Events ({{ span.events.length }})
            </v-tab>
            <v-tab v-if="span.groupId" href="#pctile">Percentiles</v-tab>
          </v-tabs>

          <v-tabs-items v-model="activeTab">
            <v-tab-item value="attrs" class="pa-4">
              <AttrsTable
                :date-range="dateRange"
                :system="span.system"
                :group-id="span.groupId"
                :attrs="span.attrs"
              />
            </v-tab-item>

            <v-tab-item value="dbStatement">
              <PrismCode :code="dbStatement" language="sql" />
            </v-tab-item>
            <v-tab-item value="dbStatementPretty">
              <PrismCode :code="dbStatementPretty" language="sql" />
            </v-tab-item>

            <v-tab-item value="excStacktrace">
              <PrismCode :code="excStacktrace" />
            </v-tab-item>

            <v-tab-item value="events">
              <EventPanels :date-range="dateRange" :events="span.events" />
            </v-tab-item>

            <v-tab-item value="pctile" class="pa-4">
              <LoadPctileChart :axios-params="axiosParams" />
            </v-tab-item>
          </v-tabs-items>
        </v-sheet>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { format } from 'sql-formatter'
import { defineComponent, ref, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor } from '@/use/uql'

// Components
import NewMonitorMenu from '@/tracing/NewMonitorMenu.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'
import AttrsTable from '@/tracing/AttrsTable.vue'
import EventPanels from '@/tracing/EventPanels.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'
import { spanName, eventOrSpanName, Span } from '@/models/span'

export default defineComponent({
  name: 'SpanCard',
  components: {
    NewMonitorMenu,
    AttrsTable,
    EventPanels,
    LoadPctileChart,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    span: {
      type: Object as PropType<Span>,
      required: true,
    },
    fluid: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const route = useRoute()
    const activeTab = ref('attrs')

    const isEvent = computed((): boolean => {
      return isEventSystem(props.span.system)
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.span.system,
        group_id: props.span.groupId,
      }
    })

    const dbStatement = computed((): string => {
      return props.span.attrs[AttrKey.dbStatement] ?? ''
    })

    const dbStatementPretty = computed((): string => {
      try {
        return format(dbStatement.value)
      } catch (err) {
        return ''
      }
    })

    const excStacktrace = computed((): string => {
      return props.span.attrs[AttrKey.exceptionStacktrace] ?? ''
    })

    const traceRoute = computed(() => {
      if (props.span.standalone) {
        return null
      }
      if (route.value.name === 'TraceShow' && route.value.params.traceId === props.span.traceId) {
        return null
      }

      return {
        name: 'TraceShow',
        params: {
          traceId: props.span.traceId,
        },
      }
    })

    const spanGroupRoute = computed(() => {
      switch (route.value.name) {
        case 'SpanList':
        case 'EventList':
        case 'SpanGroupList':
        case 'EventGroupList':
          return undefined
      }

      return {
        name: isEvent.value ? 'EventList' : 'SpanList',
        query: {
          ...props.dateRange.queryParams(),
          system: props.span.system,
          query: createUqlEditor()
            .exploreAttr(AttrKey.spanGroupId, isEvent.value)
            .where(AttrKey.spanGroupId, '=', props.span.groupId)
            .toString(),
        },
      }
    })

    return {
      AttrKey,
      activeTab,
      isEvent,

      axiosParams,

      dbStatement,
      dbStatementPretty,
      excStacktrace,

      spanGroupRoute,
      traceRoute,

      spanName,
      eventOrSpanName,
    }
  },
})
</script>

<style lang="scss" scoped></style>