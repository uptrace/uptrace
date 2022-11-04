<template>
  <XPlaceholder>
    <UptraceQuery :uql="uql" class="mt-1 mb-3">
      <SpanQueryBuilder
        :uql="uql"
        :systems="systems"
        :axios-params="axiosParams"
        :agg-disabled="['EventGroupList', 'SpanGroupList'].indexOf($route.name) === -1"
        @click:reset="resetQuery"
      />
    </UptraceQuery>

    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>{{ $route.name === 'SpanList' ? 'Spans' : 'Events' }}</span>
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
                :events-mode="eventsMode"
                :loading="spans.loading"
                :spans="spans.items"
                :order="spans.order"
                :pager="spans.pager"
                :show-system="showSystem"
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
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useSpans } from '@/tracing/use-spans'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanQueryBuilder from '@/tracing/query/SpanQueryBuilder.vue'
import SpansTable from '@/tracing/SpansTable.vue'
import { SpanChip } from '@/tracing/SpanChips.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'

// Utilities
import { isDummySystem, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'TracingSpans',
  components: { UptraceQuery, SpanQueryBuilder, SpansTable, LoadPctileChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const spans = useSpans(
      () => {
        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/spans`,
          params: props.axiosParams,
        }
      },
      {
        order: {
          syncQuery: true,
        },
      },
    )

    const showSystem = computed(() => {
      if (route.value.params.eventSystem) {
        return false
      }

      const systems = props.systems.activeSystem
      if (systems.length > 1) {
        return true
      }
      if (systems.length === 1) {
        return isDummySystem(systems[0])
      }
      return false
    })

    watch(
      () => props.eventsMode,
      (isEvent) => {
        spans.order.column = isEvent ? AttrKey.spanTime : AttrKey.spanDuration
        spans.order.desc = true
      },
      { immediate: true },
    )

    watch(
      () => spans.queryParts,
      (queryParts) => {
        if (queryParts) {
          props.uql.syncParts(queryParts)
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
      props.uql.query = props.query
    }

    function onChipClick(chip: SpanChip) {
      const editor = props.uql.createEditor()
      editor.where(chip.key, '=', chip.value)
      props.uql.commitEdits(editor)
    }

    return {
      route,
      spans,
      showSystem,

      resetQuery,
      onChipClick,
    }
  },
})
</script>

<style lang="scss" scoped></style>
