<template>
  <XPlaceholder>
    <UptraceQuery :uql="uql" class="mt-1 mb-3">
      <SpanFilters
        :uql="uql"
        :systems="systems"
        :axios-params="axiosParams"
        :agg-disabled="['LogGroupList', 'SpanGroupList'].indexOf($route.name) === -1"
        @click:reset="resetQuery"
      />
    </UptraceQuery>

    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>{{ $route.name === 'SpanList' ? 'Spans' : 'Logs' }}</span>
            </v-toolbar-title>

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <strong><XNum :value="spans.pager.numItem" verbose /></strong> spans
            </div>
          </v-toolbar>

          <v-row class="px-4 pb-4">
            <v-col>
              <LoadPctileChart :axios-params="axiosParams" class="pa-4" />

              <SpansTable
                :date-range="dateRange"
                :loading="spans.loading"
                :spans="spans.items"
                :is-event="systems.isEvent"
                :order="spans.order"
                :pager="spans.pager"
                :system="systems.activeSystem"
                @click:chip="onChipClick"
              />
            </v-col>
          </v-row>
        </v-card>

        <XPagination :pager="spans.pager" />
      </v-col>
    </v-row>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseEnvs, UseServices } from '@/tracing/use-sticky-filters'
import { UseSystems } from '@/use/systems'
import { useUql } from '@/use/uql'
import { useSpans } from '@/tracing/use-spans'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanFilters from '@/tracing/uql/SpanFilters.vue'
import SpansTable from '@/tracing/SpansTable.vue'
import { SpanChip } from '@/tracing/SpanChips.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'SpanList',
  components: { UptraceQuery, SpanFilters, SpansTable, LoadPctileChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    envs: {
      type: Object as PropType<UseEnvs>,
      required: true,
    },
    services: {
      type: Object as PropType<UseServices>,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const uql = useUql({
      syncQuery: true,
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
        ...props.envs.axiosParams(),
        ...props.services.axiosParams(),
        system: props.systems.activeSystem,
      }
    })

    const spans = useSpans(
      () => {
        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/spans`,
          params: axiosParams.value,
        }
      },
      {
        order: {
          syncQuery: true,
        },
      },
    )

    watch(
      () => props.systems.isEvent,
      (isEvent) => {
        spans.order.column = isEvent ? xkey.spanTime : xkey.spanDuration
        spans.order.desc = true
      },
      { immediate: true },
    )

    watch(
      () => spans.queryParts,
      (queryParts) => {
        if (queryParts) {
          uql.syncParts(queryParts)
        }
      },
    )

    watch(
      () => props.query,
      () => {
        if (!route.value.query.query) {
          resetQuery()
        }
      },
      { immediate: true },
    )

    function resetQuery() {
      uql.query = props.query
    }

    function onChipClick(chip: SpanChip) {
      const editor = uql.createEditor()
      editor.where(chip.key, '=', chip.value)
      uql.commitEdits(editor)
    }

    return {
      route,
      uql,
      axiosParams,
      spans,

      resetQuery,
      onChipClick,
    }
  },
})
</script>

<style lang="scss" scoped></style>
