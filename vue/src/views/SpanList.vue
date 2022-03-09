<template>
  <XPlaceholder>
    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>Spans</span>
            </v-toolbar-title>

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <strong><XNum :value="spans.pager.numItem" verbose /></strong> spans
            </div>
          </v-toolbar>

          <v-row class="px-4 pb-4">
            <v-col>
              <SpansTable
                :date-range="dateRange"
                :loading="spans.loading"
                :spans="spans.items"
                :order="spans.order"
                :pager="spans.pager"
                :system="systems.activeValue"
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
import { defineComponent, watch, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/use/systems'
import { UseUql } from '@/use/uql'
import { useSpans } from '@/use/spans'

// Components
import SpansTable from '@/components/SpansTable.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'SpanList',
  components: { SpansTable },

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
    const { route } = useRouter()

    const spans = useSpans(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/tracing/${projectId}/spans`,
        params: props.axiosParams,
      }
    })

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
          props.uql.syncParts(queryParts)
        }
      },
    )

    return { spans }
  },
})
</script>
