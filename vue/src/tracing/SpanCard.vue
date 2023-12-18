<template>
  <div>
    <PageToolbar :loading="loading" :fluid="fluid">
      <v-breadcrumbs :items="breadcrumbs" divider=">" large></v-breadcrumbs>

      <v-spacer />

      <FixedDateRangePicker :date-range="dateRange" :around="span.time" />
    </PageToolbar>

    <v-container :fluid="fluid" class="py-4">
      <SpanBodyCard :date-range="dateRange" :span="span" @find:span="$emit('find:span', $event)" />
    </v-container>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor } from '@/use/uql'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import SpanBodyCard from '@/tracing/SpanBodyCard.vue'

// Utitlies
import { Span } from '@/models/span'
import { isSpanSystem, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpanCard',
  components: { FixedDateRangePicker, SpanBodyCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    span: {
      type: Object as PropType<Span>,
      required: true,
    },
    fluid: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: props.span.system,
        to: {
          name: 'SpanGroupList',
          params: { projectId: props.span.projectId },
          query: {
            ...props.dateRange.queryParams(),
            system: props.span.system,
          },
        },
        exact: true,
      })

      bs.push({
        text: truncate(props.span.displayName, { length: 40 }),
        to: {
          name: 'SpanList',
          params: { projectId: props.span.projectId },
          query: {
            ...props.dateRange.queryParams(),
            system: props.span.system,
            query: createQueryEditor()
              .exploreAttr(AttrKey.spanGroupId, isSpanSystem(props.span.system))
              .where(AttrKey.spanGroupId, '=', props.span.groupId)
              .toString(),
          },
        },
        exact: true,
      })

      if (!props.span.standalone && props.span.traceId) {
        bs.push({
          text: props.span.traceId,
          to: {
            name: 'TraceShow',
            params: {
              projectId: props.span.projectId,
              traceId: props.span.traceId,
            },
          },
          exact: true,
        })
      }

      bs.push({
        text: 'Span',
        to: {
          name: 'SpanShow',
          params: {
            projectId: props.span.projectId,
            traceId: props.span.traceId,
            spanId: props.span.id,
          },
        },
        exact: true,
      })

      return bs
    })

    return { breadcrumbs }
  },
})
</script>

<style lang="scss" scoped></style>
