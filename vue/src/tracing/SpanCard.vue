<template>
  <div>
    <PageToolbar :fluid="fluid">
      <v-breadcrumbs :items="meta.breadcrumbs" divider=">" large>
        <template #item="{ item }">
          <v-breadcrumbs-item :to="item.to" :exact="item.exact">
            {{ item.text }}
          </v-breadcrumbs-item>
        </template>
      </v-breadcrumbs>

      <v-spacer />

      <FixedDateRangePicker :date-range="dateRange" :around="span.time" />
    </PageToolbar>

    <v-container :fluid="fluid" class="py-4">
      <v-row class="px-4 text-subtitle-1">
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

      <v-row align="end" class="px-4 text-subtitle-2 text-center">
        <v-col v-if="span.attrs[xkey.deploymentEnvironment]" cols="auto">
          <div class="grey--text font-weight-regular">Env</div>
          <div>{{ span.attrs[xkey.deploymentEnvironment] }}</div>
        </v-col>

        <v-col v-if="span.attrs[xkey.serviceName]" cols="auto">
          <div class="grey--text font-weight-regular">Service</div>
          <div>{{ span.attrs[xkey.serviceName] }}</div>
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
            <v-btn v-if="meta.traceRoute" depressed small :to="meta.traceRoute" exact>
              View trace
            </v-btn>
            <v-btn
              v-if="$route.name !== groupListRoute"
              depressed
              small
              :to="meta.groupRoute"
              exact
              class="ml-2"
            >
              View group
            </v-btn>
          </div>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet outlined rounded="lg">
            <v-tabs v-model="activeTab" background-color="transparent" class="light-blue lighten-5">
              <v-tab href="#attrs">Attrs</v-tab>
              <v-tab v-if="dbStatement" href="#dbStatement">SQL</v-tab>
              <v-tab v-if="dbStatementPretty" href="#dbStatementPretty">SQL pretty</v-tab>
              <v-tab v-if="excStacktrace" href="#excStacktrace">Stacktrace</v-tab>
              <v-tab v-if="span.events && span.events.length" href="#events">
                Events ({{ span.events.length }})
              </v-tab>
              <v-tab v-if="span.groupId" href="#pctile">Percentiles</v-tab>
            </v-tabs>

            <v-tabs-items v-model="activeTab">
              <v-tab-item value="attrs" class="pa-4">
                <AttrsTable :date-range="dateRange" :span="span" />
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
    </v-container>
  </div>
</template>

<script lang="ts">
import { format } from 'sql-formatter'
import { defineComponent, ref, computed, proxyRefs, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor } from '@/use/uql'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'
import AttrsTable from '@/tracing/AttrsTable.vue'
import EventPanels from '@/tracing/EventPanels.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { spanName, eventOrSpanName, Span } from '@/models/span'

interface Props {
  dateRange: UseDateRange
  span: Span
  groupListRoute: string
}

export default defineComponent({
  name: 'SpanCard',
  components: {
    FixedDateRangePicker,
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
    spanListRoute: {
      type: String,
      default: 'SpanList',
    },
    groupListRoute: {
      type: String,
      default: 'SpanGroupList',
    },
  },

  setup(props) {
    const activeTab = ref('attrs')

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.span.system,
        group_id: props.span.groupId,
      }
    })

    const dbStatement = computed((): string => {
      return props.span.attrs[xkey.dbStatement] ?? ''
    })

    const dbStatementPretty = computed((): string => {
      try {
        return format(dbStatement.value)
      } catch (err) {
        return ''
      }
    })

    const excStacktrace = computed((): string => {
      return props.span.attrs[xkey.exceptionStacktrace] ?? ''
    })

    return {
      xkey,
      meta: useMeta(props),
      activeTab,

      axiosParams,

      dbStatement,
      dbStatementPretty,
      excStacktrace,

      spanName,
    }
  },
})

function useMeta(props: Props) {
  const { route } = useRouter()

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

  const groupRoute = computed(() => {
    return {
      name: props.spanListRoute,
      query: {
        ...props.dateRange.queryParams(),
        system: props.span.system,
        query: createUqlEditor()
          .exploreAttr(xkey.spanGroupId)
          .where(xkey.spanGroupId, '=', props.span.groupId)
          .toString(),
      },
    }
  })

  const breadcrumbs = computed(() => {
    const bs: any[] = []

    bs.push({
      text: props.span.system,
      to: {
        name: props.groupListRoute,
        query: {
          ...props.dateRange.queryParams(),
          system: props.span.system,
        },
      },
      exact: true,
    })

    bs.push({
      text: eventOrSpanName(props.span, 50),
      to: {
        name: props.spanListRoute,
        query: {
          ...props.dateRange.queryParams(),
          system: props.span.system,
          query: createUqlEditor()
            .exploreAttr(xkey.spanGroupId)
            .where(xkey.spanGroupId, '=', props.span.groupId)
            .toString(),
        },
      },
      exact: true,
    })

    if (props.span.standalone) {
      bs.push({
        text: props.span.traceId,
        to: {
          name: 'SpanShow',
          params: {
            traceId: props.span.traceId,
            spanId: props.span.id,
          },
        },
        exact: true,
      })
    } else {
      bs.push({
        text: props.span.traceId,
        to: {
          name: 'TraceShow',
          params: {
            traceId: props.span.traceId,
          },
        },
        exact: true,
      })

      bs.push({
        text: 'Span',
        to: {
          name: 'SpanShow',
          params: {
            traceId: props.span.traceId,
            spanId: props.span.id,
          },
        },
        exact: true,
      })
    }

    return bs
  })

  return proxyRefs({ groupRoute, traceRoute, breadcrumbs })
}
</script>

<style lang="scss" scoped></style>
