<template>
  <div v-frag>
    <tr class="cursor-pointer" @click="expanded = !expanded">
      <td>
        <span class="mr-3 text--secondary"><XDate :date="timestamp" format="full" /></span>
        <span>{{ line }}</span>
      </td>
    </tr>
    <tr v-if="expanded" class="v-data-table__expanded v-data-table__expanded__content">
      <td colspan="99" class="px-6 pt-3 pb-4">
        <LogLabelsTable :labels="labels" :detected-labels="detectedLabels" />
      </td>
    </tr>
  </div>
</template>

<script lang="ts">
import { parse as parseLogfmt } from 'logfmt'
import { defineComponent, shallowRef, computed, PropType } from '@vue/composition-api'

// Components
import LogLabelsTable from '@/components/loki/LogLabelsTable.vue'

export default defineComponent({
  name: 'LogTableRow',
  components: { LogLabelsTable },

  props: {
    labels: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    timestamp: {
      type: String,
      required: true,
    },
    line: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const expanded = shallowRef(false)

    const detectedLabels = computed(() => {
      return parseLogfmt(props.line)
    })

    return { expanded, detectedLabels }
  },
})
</script>

<style lang="scss" scoped></style>
