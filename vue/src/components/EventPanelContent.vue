<template>
  <div>
    <v-row align="end" class="text-subtitle-2 text-center">
      <v-col v-if="event.attrs[xkey.serviceName]" cols="auto">
        <div class="grey--text font-weight-regular">Service</div>
        <div>{{ event.attrs[xkey.serviceName] }}</div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <XDate :date="event.time" format="full" />
      </v-col>

      <v-col cols="auto">
        <v-btn depressed small :to="groupRoute" exact class="ml-2">View group</v-btn>
        <slot name="append-action"></slot>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <PctileChart :loading="percentiles.loading" :data="percentiles.data" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <AttrTable :date-range="dateRange" :span="event" />
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

// Components
import PctileChart from '@/components/PctileChart.vue'
import AttrTable from '@/components/AttrTable.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { Span } from '@/models/span'

export default defineComponent({
  name: 'EventPanelContent',
  components: { PctileChart, AttrTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    event: {
      type: Object as PropType<Span>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const percentiles = usePercentiles(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/tracing/${projectId}/percentiles`,
        params: {
          ...props.dateRange.axiosParams(),
          system: props.event.system,
          group_id: props.event.groupId,
        },
      }
    })

    const groupRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...props.dateRange.queryParams(),
          system: props.event.system,
          where: `${xkey.spanGroupId} = "${props.event.groupId}"`,
        },
      }
    })
    return {
      xkey,

      percentiles,
      groupRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
