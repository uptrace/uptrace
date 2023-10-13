<template>
  <div>
    <v-row align="end" class="text-subtitle-2 text-center">
      <v-col v-if="event.attrs[AttrKey.serviceName]" cols="auto">
        <div class="grey--text font-weight-regular">Service</div>
        <div>{{ event.attrs[AttrKey.serviceName] }}</div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <XDate :date="event.time" format="full" />
      </v-col>

      <v-col cols="auto">
        <v-btn v-if="groupRoute" depressed small :to="groupRoute" exact class="ml-2"
          >View group</v-btn
        >
        <NewMonitorMenu
          v-if="event.groupId"
          :name="`${event.system} > ${event.name}`"
          :where="`where ${AttrKey.spanGroupId} = '${event.groupId}'`"
          events-mode
          verbose
          class="ml-2"
        />
        <slot name="append-action"></slot>
      </v-col>
    </v-row>

    <v-row v-if="event.groupId">
      <v-col>
        <PctileChart
          :annotations="annotations"
          :loading="percentiles.loading"
          :data="percentiles.data"
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
import { usePercentiles } from '@/use/percentiles'
import { createUqlEditor } from '@/use/uql'
import { Annotation } from '@/org/use-annotations'

// Components
import PctileChart from '@/components/PctileChart.vue'
import NewMonitorMenu from '@/tracing/NewMonitorMenu.vue'
import SpanAttrs from '@/tracing/SpanAttrs.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { SpanEvent } from '@/models/span'

export default defineComponent({
  name: 'EventPanelContent',
  components: { PctileChart, NewMonitorMenu, SpanAttrs },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
    event: {
      type: Object as PropType<SpanEvent>,
      required: true,
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
        url: `/api/v1/tracing/${projectId}/percentiles`,
        params: {
          ...props.dateRange.axiosParams(),
          system: props.event.system,
          group_id: props.event.groupId,
        },
      }
    })

    const groupRoute = computed(() => {
      if (!props.event.groupId) {
        return undefined
      }
      return {
        name: 'SpanList',
        query: {
          ...props.dateRange.queryParams(),
          system: props.event.system,
          query: createUqlEditor()
            .exploreAttr(AttrKey.spanGroupId, true)
            .where(AttrKey.spanGroupId, '=', props.event.groupId)
            .toString(),
        },
      }
    })

    return {
      AttrKey,

      percentiles,
      groupRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
