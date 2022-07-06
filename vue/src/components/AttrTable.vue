<template>
  <div>
    <v-simple-table class="v-data-table--border table-scroll-target">
      <colgroup>
        <col />
        <col />
      </colgroup>

      <thead class="v-data-table-header">
        <tr>
          <th>Key</th>
          <th class="target">Value</th>
        </tr>
      </thead>

      <tbody v-if="!attrKeys.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99">There are no attributes matching the filters.</td>
        </tr>
      </tbody>

      <tbody>
        <tr v-for="key in attrKeys" :key="key">
          <th>{{ key }}</th>
          <td>
            <KeyValueFilterLink
              :date-range="dateRange"
              :name="key"
              :value="span.attrs[key]"
              :project-id="span.projectId"
              :system="span.system"
              :group-id="span.groupId"
              :is-event="isEvent"
            />
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Utilities
import { xkey, isEventSystem } from '@/models/otelattr'
import { Span } from '@/models/span'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import KeyValueFilterLink from '@/components/KeyValueFilterLink.vue'

export default defineComponent({
  name: 'AttrTable',
  components: { KeyValueFilterLink },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    span: {
      type: Object as PropType<Span>,
      required: true,
    },
  },

  setup(props) {
    const attrKeys = computed((): string[] => {
      const keys = Object.keys(props.span.attrs)
      keys.sort()
      return keys
    })

    const isEvent = computed((): boolean => {
      return isEventSystem(props.span.system)
    })

    return {
      xkey,
      attrKeys,
      isEvent,
    }
  },
})
</script>

<style lang="scss" scoped></style>
