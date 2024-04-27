<template>
  <v-container fluid>
    <v-row
      align="end"
      :dense="$vuetify.breakpoint.mdAndDown"
      class="px-2 text-subtitle-2 text-center"
    >
      <v-col cols="auto">
        <div class="grey--text font-weight-regular">{{ AttrKey.spanGroupId }}</div>
        <div>{{ groupId }}</div>
      </v-col>
      <v-col v-if="group.firstSeenAt" cols="auto">
        <div class="grey--text font-weight-regular">First seen at</div>
        <div><DateValue :value="group.firstSeenAt" format="short" /></div>
      </v-col>
      <v-col v-if="group.lastSeenAt" cols="auto">
        <div class="grey--text font-weight-regular">Last seen at</div>
        <div><DateValue :value="group.lastSeenAt" format="short" /></div>
      </v-col>
      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Rate</div>
        <div><NumValue :value="group.getMetric('per_min(sum(_count))')" /> / min</div>
      </v-col>
      <v-col v-if="isSpan" cols="auto">
        <div class="grey--text font-weight-regular">Err rate</div>
        <div>
          <PctValue :a="group.getMetric('_error_count')" :b="group.getMetric('_count')" />
        </div>
      </v-col>
      <v-col v-if="isSpan" cols="auto">
        <div class="grey--text font-weight-regular">P50</div>
        <div><DurationValue :value="group.getMetric('p50(_duration)')" /></div>
      </v-col>
      <v-col v-if="isSpan" cols="auto">
        <div class="grey--text font-weight-regular">P90</div>
        <div><DurationValue :value="group.getMetric('p90(_duration)')" /></div>
      </v-col>
      <v-col v-if="isSpan" cols="auto">
        <div class="grey--text font-weight-regular">P99</div>
        <div><DurationValue :value="group.getMetric('p99(_duration)')" /></div>
      </v-col>
      <v-col v-if="isSpan" cols="auto">
        <div class="grey--text font-weight-regular">Max</div>
        <div><DurationValue :value="group.getMetric('max(_duration)')" /></div>
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-expansion-panels :value="[0]" multiple>
          <v-expansion-panel>
            <v-expansion-panel-header>Percentiles chart</v-expansion-panel-header>
            <v-expansion-panel-content>
              <PercentilesChartLazy :axios-params="axiosParams" :annotations="annotations" />
            </v-expansion-panel-content>
          </v-expansion-panel>
          <v-expansion-panel v-if="isSpan">
            <v-expansion-panel-header>
              <div class="d-flex align-center">
                <div style="min-width: 280px">Slowest spans</div>
                <v-btn
                  :to="{
                    name: 'SpanList',
                    query: {
                      system: system,
                      query: slowestSpansQuery,
                      sort_by: AttrKey.spanDuration,
                    },
                  }"
                  exact
                  small
                  outlined
                  plain
                  >Open in explorer</v-btn
                >
              </div>
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <UqlCardReadonly :query="slowestSpansQuery" class="mb-4" />

              <PagedSpansCardLazy
                :date-range="dateRange"
                :is-span="isSpan"
                :axios-params="slowestSpansAxiosParams"
              />
            </v-expansion-panel-content>
          </v-expansion-panel>
          <v-expansion-panel v-if="isSpan">
            <v-expansion-panel-header>
              <div class="d-flex align-center">
                <div style="min-width: 280px">Spans with .status_code = 'error'</div>
                <v-btn
                  :to="{
                    name: 'SpanList',
                    query: {
                      system: system,
                      query: failedSpansQuery,
                      sort_by: AttrKey.spanDuration,
                    },
                  }"
                  exact
                  small
                  outlined
                  plain
                  >Open in explorer</v-btn
                >
              </div>
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <UqlCardReadonly :query="failedSpansQuery" class="mb-4" />

              <PagedSpansCardLazy
                :date-range="dateRange"
                :is-span="isSpan"
                :axios-params="failedSpansAxiosParams"
              />
            </v-expansion-panel-content>
          </v-expansion-panel>
          <v-expansion-panel v-if="isSpan">
            <v-expansion-panel-header>
              <div class="d-flex align-center">
                <div style="min-width: 280px">Status messages</div>
                <v-btn
                  :to="{
                    name: 'SpanGroupList',
                    query: { system: system, query: statusMessagesQuery },
                  }"
                  exact
                  small
                  outlined
                  plain
                  >Open in explorer</v-btn
                >
              </div>
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <UqlCardReadonly :query="statusMessagesQuery" class="mb-4" />

              <PagedGroupsCardLazy
                :date-range="dateRange"
                :systems="[system]"
                :query="statusMessagesQuery"
              />
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor, injectQueryStore } from '@/use/uql'
import { useGroup } from '@/tracing/use-groups'

// Components
import PercentilesChartLazy from '@/components/PercentilesChartLazy.vue'
import UqlCardReadonly from '@/components/UqlCardReadonly.vue'
import PagedGroupsCardLazy from '@/tracing/PagedGroupsCardLazy.vue'

// Misc
import { isSpanSystem, SystemName, AttrKey } from '@/models/otel'
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'GroupInfoCard',
  components: {
    PercentilesChartLazy,
    UqlCardReadonly,
    PagedGroupsCardLazy,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      required: true,
    },
    annotations: {
      type: Array,
      default: () => [],
    },
  },

  setup(props) {
    const route = useRoute()
    const { where } = injectQueryStore()

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        group_id: props.groupId,
        query: where.value,
      }
    })

    const group = useGroup(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/groups/${props.groupId}`,
        params: axiosParams.value,
      }
    })

    const isSpan = computed((): boolean => {
      return isSpanSystem(props.system)
    })

    const slowestSpansQuery = computed(() => {
      return createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId, true)
        .where(AttrKey.spanGroupId, '=', props.groupId)
        .add(where.value)
        .toString()
    })
    const slowestSpansAxiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        query: slowestSpansQuery.value,
      }
    })

    const failedSpansQuery = computed(() => {
      return createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId, true)
        .where(AttrKey.spanGroupId, '=', props.groupId)
        .where(AttrKey.spanStatusCode, '=', 'error')
        .add(where.value)
        .toString()
    })
    const failedSpansAxiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        query: failedSpansQuery.value,
      }
    })

    const statusMessagesQuery = computed(() => {
      return createQueryEditor()
        .exploreAttr(AttrKey.spanStatusMessage, true)
        .where(AttrKey.spanGroupId, '=', props.groupId)
        .where(AttrKey.spanStatusMessage, '!=', '')
        .add(where.value)
        .toString()
    })

    return {
      SystemName,
      AttrKey,
      Unit,

      axiosParams,
      group,
      isSpan,
      slowestSpansQuery,
      slowestSpansAxiosParams,
      failedSpansQuery,
      failedSpansAxiosParams,
      statusMessagesQuery,
    }
  },
})
</script>
<style lang="scss" scoped></style>
