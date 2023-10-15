<template>
  <div v-frag>
    <template v-if="trace.error">
      <TraceError v-if="trace.error" :error="trace.error" />
    </template>

    <template v-else-if="!trace.root">
      <v-container :fluid="$vuetify.breakpoint.mdAndDown">
        <v-skeleton-loader type="article, image, table" />
      </v-container>
    </template>

    <div v-else>
      <PageToolbar :loading="trace.loading" :fluid="$vuetify.breakpoint.lgAndDown">
        <v-breadcrumbs :items="meta.breadcrumbs" divider=">" large>
          <template #item="{ item }">
            <v-breadcrumbs-item :to="item.to" :exact="item.exact">
              {{ item.text }}
            </v-breadcrumbs-item>
          </template>
        </v-breadcrumbs>

        <v-spacer />

        <FixedDateRangePicker
          v-if="trace.root"
          :date-range="dateRange"
          :around="trace.root.time"
          show-reload
        />
      </PageToolbar>

      <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="py-4">
        <v-row v-if="trace.root" class="px-2 text-body-2">
          <v-col class="word-break-all">
            {{ trace.root.displayName }}
          </v-col>
        </v-row>

        <v-row v-if="trace.root" align="end" class="px-2 text-subtitle-2 text-center">
          <v-col v-if="trace.root.kind" cols="auto">
            <div class="grey--text font-weight-regular">Kind</div>
            <div>{{ trace.root.kind }}</div>
          </v-col>

          <v-col v-if="trace.root.statusCode" cols="auto">
            <div class="grey--text font-weight-regular">Status</div>
            <div :class="{ 'error--text': trace.root.statusCode === 'error' }">
              {{ trace.root.statusCode }}
            </div>
          </v-col>

          <v-col cols="auto">
            <div class="grey--text font-weight-regular">Time</div>
            <XDate :date="trace.root.time" format="full" />
          </v-col>

          <v-col cols="auto">
            <div class="grey--text font-weight-regular">Duration</div>
            <XDuration :duration="trace.root.duration" fixed />
          </v-col>

          <v-col cols="auto">
            <v-btn v-if="groupRoute" depressed small :to="groupRoute" exact>View group</v-btn>
            <v-btn
              v-if="exploreTraceRoute"
              depressed
              small
              :to="exploreTraceRoute"
              exact
              class="ml-2"
              >Explore trace</v-btn
            >
          </v-col>
        </v-row>

        <v-row v-if="trace.root && trace.root.groupId">
          <v-col>
            <v-card outlined rounded="lg">
              <v-card-text>
                <LoadPctileChart :axios-params="axiosParams" :annotations="annotations.items" />
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <SpanSystemBarChart :systems="trace.coloredSystems" />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <TraceTabs :date-range="dateRange" :trace="trace" :annotations="annotations.items" />
          </v-col>
        </v-row>
      </v-container>
    </div>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, watch, proxyRefs } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useDateRange, UseDateRange } from '@/use/date-range'
import { useTrace, UseTrace } from '@/tracing/use-trace'
import { createUqlEditor } from '@/use/uql'
import { useAnnotations } from '@/org/use-annotations'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'
import SpanSystemBarChart from '@/components/SpanSystemBarChart.vue'
import TraceTabs from '@/tracing/TraceTabs.vue'
import TraceError from '@/tracing/TraceError.vue'

// Utilities
import { AttrKey, SystemName } from '@/models/otel'

export default defineComponent({
  name: 'TraceShow',
  components: {
    FixedDateRangePicker,
    LoadPctileChart,
    SpanSystemBarChart,
    TraceTabs,
    TraceError,
  },

  setup() {
    useTitle('View trace')
    const dateRange = useDateRange()
    const trace = useTrace()

    const annotations = useAnnotations(() => {
      return {
        ...dateRange.axiosParams(),
      }
    })

    const axiosParams = computed(() => {
      if (!trace.root) {
        return
      }
      return {
        ...dateRange.axiosParams(),
        system: trace.root.system,
        group_id: trace.root.groupId,
      }
    })

    const groupRoute = computed(() => {
      if (!trace.root) {
        return
      }

      return {
        name: 'SpanList',
        query: {
          ...dateRange.queryParams(),
          system: trace.root.system,
          query: createUqlEditor()
            .exploreAttr(AttrKey.spanGroupId)
            .where(AttrKey.spanGroupId, '=', trace.root.groupId)
            .toString(),
        },
      }
    })

    const exploreTraceRoute = computed(() => {
      if (!trace.root) {
        return
      }

      return {
        name: 'SpanGroupList',
        query: {
          ...dateRange.queryParams(),
          system: SystemName.SpansAll,
          query: [
            `where ${AttrKey.spanTraceId} = ${trace.root.traceId}`,
            `group by ${AttrKey.spanGroupId}`,
            AttrKey.spanCount,
            AttrKey.spanErrorCount,
            `{p50,p90,p99,sum}(${AttrKey.spanDuration})`,
          ].join(' | '),
          plot: null,
        },
      }
    })

    watch(
      () => trace.root,
      (span) => {
        if (span) {
          useTitle(span.name)
        }
      },
    )

    return {
      dateRange,
      trace,
      annotations,
      meta: useMeta(dateRange, trace),

      axiosParams,
      groupRoute,
      exploreTraceRoute,
    }
  },
})

function useMeta(dateRange: UseDateRange, trace: UseTrace) {
  const breadcrumbs = computed(() => {
    const root = trace.root
    const bs: any[] = []

    if (!root) {
      return bs
    }

    if (root.system) {
      bs.push({
        text: root.system,
        to: {
          name: 'SpanGroupList',
          query: {
            ...dateRange.queryParams(),
            system: root.system,
          },
        },
        exact: true,
      })
    }

    if (root.system && root.groupId) {
      bs.push({
        text: truncate(root.displayName, { length: 60 }),
        to: {
          name: 'SpanList',
          query: {
            ...dateRange.queryParams(),
            system: root.system,
            query: createUqlEditor()
              .exploreAttr(AttrKey.spanGroupId)
              .where(AttrKey.spanGroupId, '=', root.groupId)
              .toString(),
          },
        },
        exact: true,
      })
    }

    bs.push({
      text: trace.id,
    })

    return bs
  })

  return proxyRefs({ breadcrumbs })
}
</script>

<style lang="scss" scoped></style>
