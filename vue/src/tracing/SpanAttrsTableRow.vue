<template>
  <div v-frag>
    <tr>
      <th class="user-select-text">{{ attrKey }}</th>
      <td class="target">
        <SpanAttrValue
          v-for="(attrValue, i) in attrValues"
          :key="i"
          :date-range="dateRange"
          :system="system"
          :group-id="groupId"
          :attr-key="attrKey"
          :attr-value="attrValue"
        />
      </td>
      <td>
        <v-btn v-if="isExpanded" icon title="Hide spans" @click.stop="expand(false)">
          <v-icon size="30">mdi-chevron-up</v-icon>
        </v-btn>
        <v-btn v-else icon title="View spans" @click.stop="expand(true)">
          <v-icon size="30">mdi-chevron-down</v-icon>
        </v-btn>
      </td>
    </tr>
    <tr v-if="isExpanded" class="v-data-table__expanded v-data-table__expanded__content">
      <td colspan="99" class="pt-4 pb-2">
        <SpanAttrValues
          :date-range="dateRange"
          :system="system"
          :group-id="groupId"
          :attr-key="attrKey"
          :attr-value="attrValue"
        />
      </td>
    </tr>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import SpanAttrValue from '@/tracing/SpanAttrValue.vue'
import SpanAttrValues from '@/tracing/SpanAttrValues.vue'

export default defineComponent({
  name: 'SpanAttrsTableRow',
  components: { SpanAttrValue, SpanAttrValues },

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
    attrKey: {
      type: String,
      required: true,
    },
    attrValue: {
      type: undefined,
      required: true,
    },
  },

  setup(props) {
    const isExpanded = shallowRef(false)

    function expand(flag: boolean) {
      isExpanded.value = flag
    }

    const attrValues = computed(() => {
      if (Array.isArray(props.attrValue)) {
        return props.attrValue
      }
      return [props.attrValue]
    })

    return { isExpanded, expand, attrValues }
  },
})
</script>

<style lang="scss" scoped></style>
