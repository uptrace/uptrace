<template>
  <GridItemCard
    :grid-item="gridItem"
    :readonly="preview"
    toolbar-color="transparent"
    class="border-bottom"
    :style="style"
    v-on="$listeners"
  >
    <v-card flat class="py-2 text-h5 text-center" :class="{ 'pa-6': preview }">
      <span v-if="gaugeHtml" v-html="gaugeHtml"></span>
      <span v-else>{{ gaugeText }}</span>
    </v-card>
  </GridItemCard>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'
import colors from 'vuetify/lib/util/colors'

// Composables
import { formatGauge } from '@/metrics/use-gauges'

// Components
import GridItemCard from '@/metrics/GridItemCard.vue'

// Misc
import { GaugeGridItem, StyledGaugeColumn, ValueMapping, MappingOp } from '@/metrics/types'
import { numVerbose } from '@/util/fmt'

export default defineComponent({
  name: 'GaugeGard',
  components: { GridItemCard },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    gridItem: {
      type: Object as PropType<GaugeGridItem>,
      required: true,
    },
    columns: {
      type: Array as PropType<StyledGaugeColumn[]>,
      required: true,
    },
    values: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    preview: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const gaugeText = computed(() => {
      return formatGauge(props.values, props.columns, props.gridItem.params.template)
    })

    // TODO: extract to separate component
    const gaugeHtml = computed(() => {
      if (!props.gridItem.params.valueMappings.length) {
        return ''
      }
      if (!props.columns.length) {
        return ''
      }
      if (props.columns.length > 1) {
        return `<span class="text-body-2 red--text">can't have multiple metrics and value mappings</span>`
      }

      const col = props.columns[0]
      const value = props.values[col.name]
      for (let mapping of props.gridItem.params.valueMappings) {
        if (mappingMatches(mapping, value)) {
          return `<span style="color: ${mapping.color}">${mapping.text}</span>`
        }
      }
      return numVerbose(value)
    })

    const style = computed(() => {
      return {
        'border-bottom-color': colors.blue.darken2,
      }
    })

    return {
      gaugeText,
      gaugeHtml,
      style,
    }
  },
})

function mappingMatches(mapping: ValueMapping, value: number): boolean {
  switch (mapping.op) {
    case MappingOp.Any:
      return true
    case MappingOp.Equal:
      return value === mapping.value
    case MappingOp.Lt:
      return value < mapping.value
    case MappingOp.Lte:
      return value <= mapping.value
    case MappingOp.Gt:
      return value > mapping.value
    case MappingOp.Gte:
      return value >= mapping.value
    default:
      return false
  }
}
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom-width: 8px;
}
</style>
