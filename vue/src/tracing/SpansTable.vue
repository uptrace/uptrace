<template>
  <div>
    <v-simple-table>
      <thead v-if="spans.length" class="v-data-table-header">
        <tr>
          <th>Span Name</th>
          <th v-if="showSystem">System</th>
          <th></th>
          <ThOrder v-for="col in columns" :key="col" :value="col" :order="order">
            <span>{{ col }}</span>
          </ThOrder>
          <ThOrder :value="AttrKey.spanTime" :order="order">Time</ThOrder>
          <ThOrder v-if="!eventsMode" :value="AttrKey.spanDuration" :order="order" align="end">
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
            <v-btn :to="{ name: 'TracingHelp' }">
              <v-icon left>mdi-help-circle-outline</v-icon>
              <span>Help</span>
            </v-btn>
          </td>
        </tr>
      </tbody>

      <tbody>
        <template v-for="(span, index) in spans">
          <tr :key="`a-${index}`" class="cursor-pointer" @click="showSpan(span)">
            <td class="word-break-all">
              {{ span.displayName }}
            </td>
            <td v-if="showSystem">
              <router-link :to="systemRoute(span)" @click.native.stop>{{
                span.system
              }}</router-link>
            </td>
            <td>
              <SpanChips
                :span="span"
                :clickable="'click:chip' in $listeners"
                @click:chip="$emit('click:chip', $event)"
              />
            </td>
            <td v-for="col in columns" :key="col">
              <AnyValue :value="span.attrs[col]" :name="col" />
            </td>
            <td class="text-no-wrap"><DateValue :value="span.time" format="relative" /></td>
            <td v-if="!eventsMode" class="text-right">
              <DurationValue :value="span.duration" fixed />
            </td>
          </tr>
        </template>
      </tbody>
    </v-simple-table>

    <v-dialog v-model="dialog" max-width="1280">
      <v-sheet>
        <SpanCard
          v-if="activeSpan"
          :date-range="internalDateRange"
          :span="activeSpan"
          fluid
          show-toolbar
        />
      </v-sheet>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, nextTick, watch, PropType } from 'vue'

// Composables
import { useDateRange, UseDateRange } from '@/use/date-range'
import { useRoute } from '@/use/router'
import { UsePager } from '@/use/pager'
import { UseOrder } from '@/use/order'
import { useAnnotations } from '@/org/use-annotations'

// Components
import SpanCard from '@/tracing/SpanCard.vue'
import ThOrder from '@/components/ThOrder.vue'
import SpanChips from '@/tracing/SpanChips.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { Span } from '@/models/span'

export default defineComponent({
  name: 'SpansTable',
  components: {
    ThOrder,
    SpanCard,
    SpanChips,
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
    eventsMode: {
      type: Boolean,
      required: true,
    },
    showSystem: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const route = useRoute()

    function systemRoute(span: Span) {
      return {
        query: {
          ...route.value.query,
          system: span.system,
        },
      }
    }

    function onSortBy() {
      nextTick(() => {
        props.order.desc = true
      })
    }

    // Dialog
    //-------

    const internalDateRange = useDateRange()
    useAnnotations(() => {
      return {
        ...internalDateRange.axiosParams(),
      }
    })

    const dialog = shallowRef(false)
    const activeSpan = shallowRef<Span>()

    watch(route, () => {
      dialog.value = false
    })

    watch(dialog, (dialog) => {
      if (dialog) {
        internalDateRange.syncWith(props.dateRange)
      }
    })

    function showSpan(span: Span) {
      activeSpan.value = span
      dialog.value = true
    }

    return {
      internalDateRange,
      AttrKey,
      dialog,
      activeSpan,

      systemRoute,
      onSortBy,

      showSpan,
    }
  },
})
</script>

<style lang="scss" scoped></style>
