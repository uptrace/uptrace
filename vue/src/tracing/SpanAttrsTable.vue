<template>
  <v-simple-table class="v-data-table--border table-scroll-target">
    <colgroup>
      <col />
      <col class="target" />
      <col />
    </colgroup>

    <thead class="v-data-table-header">
      <tr>
        <th>Key</th>
        <th class="target">Value</th>
        <th></th>
      </tr>
    </thead>

    <tbody v-if="!attrKeys.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99">There are no attributes matching the filters.</td>
      </tr>
    </tbody>

    <tbody>
      <SpanAttrsTableRow
        v-for="attrKey in attrKeys"
        :key="attrKey"
        :date-range="dateRange"
        :system="system"
        :group-id="groupId"
        :attr-key="attrKey"
        :attr-value="attrs[attrKey]"
      />
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import SpanAttrsTableRow from '@/tracing/SpanAttrsTableRow.vue'

// Utitlies
import { AttrMap } from '@/models/span'
import { AttrKey, isEventSystem } from '@/models/otel'

export default defineComponent({
  name: 'SpanAttrsTable',
  components: { SpanAttrsTableRow },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      required: true,
    },
    attrs: {
      type: Object as PropType<AttrMap>,
      required: true,
    },
    attrKeys: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },

  setup(props) {
    const isEvent = computed((): boolean => {
      return isEventSystem(props.system)
    })

    return {
      AttrKey,
      isEvent,
    }
  },
})
</script>

<style lang="scss" scoped></style>
