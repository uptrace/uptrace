<template>
  <div>
    <table class="trace-table trace-table--timeline">
      <colgroup>
        <col style="width: 50%" />
        <col style="width: 50%" />
      </colgroup>

      <thead>
        <tr>
          <th class="px-2">Operation</th>
          <th>Timeline</th>
        </tr>
      </thead>

      <tbody>
        <template v-for="span in trace.spans">
          <tr
            v-if="trace.isVisible(span)"
            :id="`span-${span.id}`"
            :key="`span-${span.id}`"
            class="cursor-pointer"
            :class="{ active: span === trace.activeSpan }"
            @click="showSpan(span)"
          >
            <td :style="{ paddingLeft: 36 + 20 * span.level + 'px' }">
              <v-btn
                icon
                :disabled="!span.children"
                class="ml-n9"
                @click.stop="trace.toggleTree(span)"
              >
                <v-icon size="28">{{
                  trace.isExpanded(span) ? 'mdi-chevron-down' : 'mdi-chevron-right'
                }}</v-icon>
              </v-btn>

              <span class="cursor-pointer span-name">{{ spanName(span, 100) }}</span>
              <SpanChips :span="span" trace-mode class="ml-2" />
            </td>

            <td class="text-caption" style="position: relative">
              <span :style="span.labelStyle">
                <span v-show="span.children && trace.isExpanded(span)">
                  <XDuration :duration="span.durationSelf" />
                  <span class="mx-1">of</span>
                </span>
                <XDuration :duration="span.duration" />
              </span>

              <span
                v-for="(bar, i) in span.bars"
                :key="i"
                :title="`${durationFixed(bar.duration)} ${spanName(span)}`"
                class="span-bar"
                :style="bar.coloredStyle"
              ></span>

              <TraceTimelineChildrenBars
                v-if="span.children"
                :trace="trace"
                :span="span"
                :children="span.children"
                :hidden="!trace.isExpanded(span)"
              />
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <v-dialog v-model="dialog" max-width="1280">
      <v-sheet>
        <SpanCard
          v-if="trace.activeSpan"
          :key="trace.activeSpan.id"
          :date-range="dateRange"
          :span="trace.activeSpan"
          fluid
        />
      </v-sheet>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { UseTrace, TraceSpan } from '@/use/trace'

// Components
import SpanCard from '@/components/SpanCard.vue'
import SpanChips from '@/components/SpanChips.vue'
import TraceTimelineChildrenBars from '@/components/TraceTimelineChildrenBars.vue'

// Utilities
import { spanName } from '@/models/span'
import { durationFixed } from '@/util/fmt/duration'

export default defineComponent({
  name: 'TraceTimeline',
  components: { SpanCard, SpanChips, TraceTimelineChildrenBars },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    trace: {
      type: Object as PropType<UseTrace>,
      required: true,
    },
  },

  setup(props) {
    const dialog = shallowRef(false)

    function showSpan(span: TraceSpan) {
      dialog.value = true
      props.trace.activeSpanId = span.id
    }

    return { dialog, showSpan, spanName, durationFixed }
  },
})
</script>

<style lang="scss" scoped>
.v-btn.v-btn--disabled .v-icon {
  cursor: initial !important;
  color: rgba(255, 255, 255, 0) !important;
}

tr.active {
  background: map-deep-get($material-theme, 'table', 'active');

  .span-name {
    font-weight: 600;
  }
}
</style>

<style lang="scss">
.trace-table {
  width: 100%;
  border-spacing: 0;
  font-size: map-deep-get($headings, 'body-2', 'size');

  & > thead > tr > th {
    text-align: left;
    font-size: 12px;
    color: rgba(
      map-get($material-theme, 'text-color'),
      map-get($material-theme, 'secondary-text-percent')
    );
  }

  & > tbody > tr {
    & > td {
      border-bottom: 1px solid
        rgba(map-get($material-theme, 'fg-color'), map-get($material-theme, 'divider-percent'));
      padding-top: 22px;
      padding-bottom: 0px;
    }

    &:hover {
      background: map-deep-get($material-theme, 'table', 'hover');
    }
  }
}

.trace-table--timeline {
  table-layout: fixed;

  .span-bar {
    position: absolute;
    bottom: -6px;
    z-index: 1;
    height: 11px;
    margin-right: 1px;
  }

  & > tbody > tr > td:first-child {
    padding-right: 8px;
  }
}

.trace-table--tree {
  & > thead > tr > th,
  & > tbody > tr > td {
    padding: 0 10px;
  }
}
</style>
