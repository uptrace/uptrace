<template>
  <v-row>
    <v-col>
      <v-card rounded="lg" outlined class="mb-4">
        <v-toolbar flat color="light-blue lighten-5">
          <slot name="search-filter" />
          <SpansTableSettings :spans="spans.items" class="ml-2" @input="tableColumns = $event" />

          <v-spacer />

          <div class="text-body-2 blue-grey--text text--darken-3">
            <strong><NumValue :value="spans.pager.numItem" verbose /></strong> spans
          </div>
        </v-toolbar>

        <v-container fluid>
          <v-row dense>
            <v-col>
              <LoadPctileChart
                :axios-params="axiosParams"
                :annotations="annotations.items"
                class="pa-4"
              />
            </v-col>
          </v-row>

          <v-row v-if="spans.error">
            <v-col>
              <ApiErrorCard :error="spans.error" />
            </v-col>
          </v-row>

          <v-row v-else dense>
            <v-col>
              <SpansTable
                :date-range="dateRange"
                :loading="spans.loading"
                :spans="spans.items"
                :order="spans.order"
                :pager="spans.pager"
                :is-span="systems.isSpan"
                :columns="tableColumns"
                :show-system="showSystem"
                @click:chip="onChipClick"
              />
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <XPagination :pager="spans.pager" />
            </v-col>
          </v-row>
        </v-container>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRouter, useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useAnnotations } from '@/org/use-annotations'
import { useSpans } from '@/tracing/use-spans'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import SpansTable from '@/tracing/SpansTable.vue'
import SpansTableSettings from '@/tracing/SpansTableSettings.vue'
import { SpanChip } from '@/tracing/SpanChips.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'

// Misc
import { isLogSystem, isGroupSystem, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'TracingSpans',
  components: {
    ApiErrorCard,
    SpansTable,
    SpansTableSettings,
    LoadPctileChart,
  },

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
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    props.dateRange.roundUp()
    const { route } = useRouter()
    const tableColumns = shallowRef<string[]>([])

    const annotations = useAnnotations(() => {
      return {
        ...props.dateRange.axiosParams(),
      }
    })

    const spans = useSpans(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/spans`,
        params: props.axiosParams,
      }
    })

    const showSystem = computed(() => {
      const systems = props.systems.activeSystems
      if (isLogSystem(...systems)) {
        return false
      }
      if (systems.length > 1) {
        return true
      }
      if (systems.length === 1) {
        return isGroupSystem(systems[0])
      }
      return false
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        props.uql.parseQueryParams(queryParams)
        spans.order.parseQueryParams(queryParams)
      },
      toQuery() {
        const queryParams: Record<string, any> = {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
          ...spans.order.queryParams(),
        }
        return queryParams
      },
    })

    watch(
      () => spans.queryInfo,
      (queryInfo) => {
        if (queryInfo) {
          props.uql.setQueryInfo(queryInfo)
        }
      },
    )

    watch(
      () => props.systems.activeSystems,
      (system, oldSystem) => {
        if (!spans.order.column || oldSystem) {
          resetOrder()
        }
      },
    )

    function resetOrder() {
      spans.order.column = ''
      spans.order.desc = true
      if (!props.systems.activeSystems) {
        return
      }
      spans.order.column = props.systems.isSpan ? AttrKey.spanDuration : AttrKey.spanTime
    }

    function onChipClick(chip: SpanChip) {
      const editor = props.uql.createEditor()
      editor.where(chip.key, '=', chip.value)
      props.uql.commitEdits(editor)
    }

    return {
      annotations,

      spans,
      showSystem,
      tableColumns,

      onChipClick,
    }
  },
})
</script>

<style lang="scss" scoped></style>
