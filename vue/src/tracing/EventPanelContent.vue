<template>
  <div>
    <v-row align="end" class="text-subtitle-2 text-center">
      <v-col v-if="event.attrs[AttrKey.serviceName]" cols="auto">
        <div class="grey--text font-weight-regular">Service</div>
        <div>{{ event.attrs[AttrKey.serviceName] }}</div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <DateValue :value="event.time" format="full" />
      </v-col>

      <v-col cols="auto">
        <v-btn v-if="spanListRoute" depressed small :to="spanListRoute" exact class="ml-2"
          >View group</v-btn
        >
        <NewMonitorMenu
          v-if="event.groupId"
          :systems="[event.system]"
          :name="`${event.system} > ${event.name}`"
          :where="`where ${AttrKey.spanGroupId} = '${event.groupId}'`"
          verbose
          class="ml-2"
        />
        <slot name="append-action"></slot>
      </v-col>
    </v-row>

    <v-row v-if="percentiles.status.hasData()">
      <v-col>
        <EventRateChart
          :loading="percentiles.loading"
          :time="percentiles.stats.time"
          :count-per-min="percentiles.stats.rate"
          :annotations="annotations"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <SpanAttrs
          :date-range="dateRange"
          :attrs="event.attrs"
          :system="event.system"
          :group-id="event.groupId"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, computed } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { usePercentiles } from '@/tracing/use-percentiles'
import { createQueryEditor } from '@/use/uql'

// Components
import EventRateChart from '@/components/EventRateChart.vue'
import SpanAttrs from '@/tracing/SpanAttrs.vue'
import NewMonitorMenu from '@/tracing/NewMonitorMenu.vue'

// Misc
import { SpanEvent } from '@/models/span'
import { AttrKey } from '@/models/otel'
import { Annotation } from '@/org/types'

export default defineComponent({
  name: 'EventPanelContent',
  components: { EventRateChart, SpanAttrs, NewMonitorMenu },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    event: {
      type: Object as PropType<SpanEvent>,
      required: true,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props) {
    const { route } = useRouter()

    const percentiles = usePercentiles(() => {
      if (!props.event.groupId) {
        return undefined
      }

      const { projectId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/percentiles`,
        params: {
          ...props.dateRange.axiosParams(),
          system: props.event.system,
          group_id: props.event.groupId,
        },
      }
    })

    const spanListRoute = computed(() => {
      if (!props.event.groupId) {
        return undefined
      }
      return {
        name: 'SpanList',
        query: {
          ...props.dateRange.queryParams(),
          system: props.event.system,
          query: createQueryEditor()
            .exploreAttr(AttrKey.spanGroupId, true)
            .where(AttrKey.spanGroupId, '=', props.event.groupId)
            .toString(),
        },
      }
    })

    return {
      AttrKey,

      percentiles,
      spanListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
