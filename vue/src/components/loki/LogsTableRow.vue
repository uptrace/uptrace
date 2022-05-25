<template>
  <div v-frag>
    <tr class="cursor-pointer" @click="expanded = !expanded">
      <td class="pa-0">
        <div class="d-flex align-center">
          <div class="mr-2 severity" :class="severityColor"></div>
          <div class="mr-1">
            <v-icon>{{ expanded ? 'mdi-chevron-down' : 'mdi-chevron-right' }}</v-icon>
          </div>
          <div class="mr-3 text--secondary"><XDate :date="timestamp" format="full" /></div>
          <div>{{ line }}</div>
        </div>
      </td>
    </tr>
    <tr v-if="expanded" class="v-data-table__expanded v-data-table__expanded__content">
      <td colspan="99" class="px-6 pt-3 pb-4">
        <v-btn
          v-if="traceId"
          :to="{ name: 'TraceFind', params: { traceId: traceId } }"
          small
          color="primary"
          class="my-2"
          >Find trace</v-btn
        >
        <LogLabelsTable
          :labels="labels"
          :detected-labels="detectedLabels"
          @click:filter="$emit('click:filter', $event)"
        />
      </td>
    </tr>
  </div>
</template>

<script lang="ts">
import { assign } from 'lodash'
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

    const detectedLabels = computed((): Record<string, string> => {
      return parseLogfmt(props.line) as Record<string, string>
    })

    const mergedLabels = computed(() => {
      const dest: Record<string, string> = {}
      return assign(dest, detectedLabels.value, props.labels)
    })

    const traceId = computed((): string => {
      for (let key of ['traceid', 'trace_id', 'traceId']) {
        const value = detectedLabels.value[key]
        if (value) {
          return value
        }
      }

      let m = props.line.match('/\b[0-9a-f]{32}\b/')
      if (m) {
        return m[0]
      }

      m = props.line.match(
        /\b[0-9a-f]{8}\b-[0-9a-f]{4}\b-[0-9a-f]{4}\b-[0-9a-f]{4}\b-[0-9a-f]{12}\b/,
      )
      if (m) {
        return m[0]
      }

      return ''
    })

    const severity = computed(() => {
      for (let key of ['log.severity', 'severity', 'level']) {
        const value = mergedLabels.value[key]
        if (value && typeof value === 'string') {
          return value.toLowerCase()
        }
      }
      return ''
    })

    const severityColor = computed(() => {
      switch (severity.value) {
        case 'info':
        case 'information':
          return 'green'
        case 'warn':
        case 'warning':
          return 'lime'
        case 'err':
        case 'error':
          return 'orange'
        case 'fatal':
          return 'red'
        default:
          return 'grey'
      }
    })

    return { expanded, detectedLabels, traceId, severity, severityColor }
  },
})
</script>

<style lang="scss" scoped>
.severity {
  height: 40px;
  width: 4px;
}
</style>
