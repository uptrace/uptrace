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

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Kind</div>
        <div>{{ span.kind }}</div>
      </v-col>

      <v-col v-if="!isEvent" cols="auto">
        <div class="grey--text font-weight-regular">Status</div>
        <div>
          <v-tooltip v-if="span.statusMessage" max-width="600" bottom>
            <template #activator="{ on, attrs }">
              <div :class="{ 'error--text': span.statusCode === 'error' }" v-bind="attrs" v-on="on">
                {{ span.statusCode }}
              </div>
            </template>
            <div>{{ span.statusMessage }}</div>
          </v-tooltip>
          <div v-else :class="{ 'error--text': span.statusCode === 'error' }">
            {{ span.statusCode }}
          </div>
        </div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <DateValue v-if="span.time" :value="span.time" format="full" />
      </v-col>

      <v-col v-if="span.duration > 0" cols="auto">
        <div class="grey--text font-weight-regular">Duration</div>
        <DurationValue :value="span.duration" fixed />
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
            :name="`${span.system} > ${span.displayName}`"
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
            <v-tab href="#group">Group</v-tab>
            <v-tab v-if="dbStatement" href="#dbStatement">SQL:raw</v-tab>
            <v-tab v-if="dbStatementPretty" href="#dbStatementPretty">SQL:pretty</v-tab>
            <v-tab v-if="excStacktrace" href="#excStacktrace">Stacktrace</v-tab>
            <v-tab v-if="span.events && span.events.length" href="#events">
              Events ({{ span.events.length }})
            </v-tab>
          </v-tabs>

          <v-tabs-items v-model="activeTab">
            <v-tab-item value="attrs" class="pa-4">
              <SpanAttrs
                :date-range="dateRange"
                :attrs="span.attrs"
                :system="span.system"
                :group-id="span.groupId"
              />
            </v-tab-item>
            <v-tab-item value="group">
              <GroupInfoCard
                :date-range="dateRange"
                :system="span.system"
                :group-id="span.groupId"
                :annotations="annotations"
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
              <EventPanels
                :date-range="dateRange"
                :events="span.events"
                :annotations="annotations"
              />
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
import { createQueryEditor, useQueryStore } from '@/use/uql'
import { injectAnnotations } from '@/org/use-annotations'

// Components
import NewMonitorMenu from '@/tracing/NewMonitorMenu.vue'
import SpanAttrs from '@/tracing/SpanAttrs.vue'
import EventPanels from '@/tracing/EventPanels.vue'
import GroupInfoCard from '@/tracing/GroupInfoCard.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'
import { spanName, Span } from '@/models/span'

export default defineComponent({
  name: 'SpanBodyCard',
  components: {
    NewMonitorMenu,
    SpanAttrs,
    EventPanels,
    // LoadPctileChart,
    GroupInfoCard,
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
    const { where } = useQueryStore()
    const activeTab = ref('attrs')

    const isEvent = computed((): boolean => {
      return isEventSystem(props.span.system)
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.span.system,
        group_id: props.span.groupId,
        query: where.value,
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
        case 'SpanGroupList':
          return undefined
      }

      return {
        name: 'SpanList',
        query: {
          ...props.dateRange.queryParams(),
          system: props.span.system,
          query: createQueryEditor()
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

      annotations: injectAnnotations(),

      axiosParams,

      dbStatement,
      dbStatementPretty,
      excStacktrace,

      spanGroupRoute,
      traceRoute,

      spanName,
    }
  },
})
</script>

<style lang="scss" scoped></style>
