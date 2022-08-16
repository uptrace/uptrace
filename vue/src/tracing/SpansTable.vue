<template>
  <div>
    <v-simple-table>
      <thead v-if="spans.length" class="v-data-table-header">
        <tr>
          <th>Span Name</th>
          <th></th>
          <ThOrder v-for="col in columns" :key="col" :value="col" :order="order">
            <span>{{ col }}</span>
          </ThOrder>
          <ThOrder :value="xkey.spanTime" :order="order">Time</ThOrder>
          <ThOrder v-if="!isEvent" :value="xkey.spanDuration" :order="order" align="end">
            <span>Dur.</span>
          </ThOrder>
        </tr>
      </thead>

      <thead v-if="loading">
        <tr class="v-data-table__progress">
          <th colspan="99" class="column">
            <v-progress-linear height="2" absolute indeterminate />
          </th>
        </tr>
      </thead>

      <tbody v-if="!spans.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99" class="py-16">
            <div class="mb-4">There are no matching spans. Try to change filters.</div>
            <v-btn :to="{ name: 'Help' }">
              <v-icon left>mdi-help-circle-outline</v-icon>
              <span>Help</span>
            </v-btn>
          </td>
        </tr>
      </tbody>

      <tbody>
        <template v-for="span in spans">
          <tr :key="`a-${span.id}`" class="cursor-pointer" @click="dialog.showSpan(span)">
            <td>
              <span>{{ eventOrSpanName(span) }}</span>
            </td>
            <td>
              <SpanChips
                :span="span"
                :clickable="'click:chip' in $listeners"
                @click:chip="$emit('click:chip', $event)"
              />
            </td>
            <td class="text-no-wrap"><XDate :date="span.time" format="relative" /></td>
            <td v-if="!isEvent" class="text-right">
              <XDuration :duration="span.duration" fixed />
            </td>
          </tr>
        </template>
      </tbody>
    </v-simple-table>

    <v-dialog v-model="dialog.dialog" max-width="1280">
      <v-sheet>
        <SpanCard
          v-if="dialog.activeSpan"
          :key="dialog.activeSpan[xkey.spanId]"
          :date-range="dateRange"
          :span="dialog.activeSpan"
          :span-list-route="spanListRoute"
          :group-list-route="groupListRoute"
          fluid
          show-toolbar
        />
      </v-sheet>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, nextTick, proxyRefs, PropType } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'
import { UsePager } from '@/use/pager'
import { UseOrder } from '@/use/order'
import { UseDateRange } from '@/use/date-range'

// Components
import ThOrder from '@/components/ThOrder.vue'
import SpanCard from '@/tracing/SpanCard.vue'
import SpanChips from '@/tracing/SpanChips.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { eventOrSpanName, Span } from '@/models/span'

export default defineComponent({
  name: 'SpansTable',
  components: {
    ThOrder,
    SpanChips,
    SpanCard,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    spans: {
      type: Array as PropType<Span[]>,
      required: true,
    },
    isEvent: {
      type: Boolean,
      default: false,
    },
    pager: {
      type: Object as PropType<UsePager>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    columns: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    spanListRoute: {
      type: String,
      required: true,
    },
    groupListRoute: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    function onSortBy() {
      nextTick(() => {
        props.order.desc = true
      })
    }

    const dialog = useDialog()

    useRouteQuery().onRouteUpdated(() => {
      dialog.close()
    })

    return {
      xkey,
      dialog,

      eventOrSpanName,
      onSortBy,
    }
  },
})

function useDialog() {
  const dialog = shallowRef(false)
  const activeSpan = shallowRef<Span>()

  function showSpan(span: Span) {
    activeSpan.value = span
    dialog.value = true
  }

  function close() {
    dialog.value = false
  }

  return proxyRefs({
    dialog,
    activeSpan,

    showSpan,
    close,
  })
}
</script>

<style lang="scss" scoped></style>
