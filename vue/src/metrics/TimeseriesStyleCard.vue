<template>
  <v-card flat :max-width="width">
    <div class="mb-4">
      <v-btn outlined @click="dialog = true">
        <v-icon size="24" :color="timeseriesStyle.color" left>mdi-circle</v-icon>
        <span>{{ timeseriesStyle.color }}</span>
      </v-btn>
    </div>

    <PanelSection v-if="showOpacity" title="Fill opacity">
      <v-slider v-model="timeseriesStyle.opacity" min="0" max="100" hide-details="auto">
        <template #prepend>{{ timeseriesStyle.opacity }}%</template>
      </v-slider>
    </PanelSection>

    <PanelSection v-if="showLineWidth" title="Line width">
      <v-slider v-model="timeseriesStyle.lineWidth" min="1" max="10" hide-details="auto">
        <template #prepend>{{ timeseriesStyle.lineWidth }}px</template>
      </v-slider>
    </PanelSection>

    <PanelSection v-if="showSymbol" title="Symbol">
      <v-select
        v-model="timeseriesStyle.symbol"
        :items="symbolItems"
        filled
        dense
        hide-details="auto"
      ></v-select>
    </PanelSection>

    <PanelSection v-if="showSymbol" title="Symbol size">
      <v-slider v-model="timeseriesStyle.symbolSize" min="1" max="16" hide-details="auto">
        <template #prepend>{{ timeseriesStyle.symbolSize }}px</template>
      </v-slider>
    </PanelSection>

    <v-dialog v-model="dialog" width="auto">
      <v-card>
        <v-container fluid>
          <v-row>
            <v-col>
              <v-color-picker
                v-model="timeseriesStyle.color"
                mode="hexa"
                show-swatches
                swatches-max-height="300"
              ></v-color-picker>
            </v-col>
          </v-row>

          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn color="primary" @click="dialog = false">OK</v-btn>
            </v-col>
          </v-row>
        </v-container>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Components
import PanelSection from '@/components/PanelSection.vue'

// Misc
import { ChartKind, TimeseriesStyle } from '@/metrics/types'

export default defineComponent({
  name: 'TimeseriesStyleCard',
  components: { PanelSection },

  props: {
    chartKind: {
      type: String as PropType<ChartKind>,
      required: true,
    },
    timeseriesStyle: {
      type: Object as PropType<TimeseriesStyle>,
      required: true,
    },
    width: {
      type: Number,
      default: 400,
    },
  },

  setup(props, ctx) {
    const dialog = shallowRef(false)

    const showOpacity = computed(() => {
      return [ChartKind.Area, ChartKind.StackedArea].indexOf(props.chartKind) >= 0
    })

    const showLineWidth = computed(() => {
      return [ChartKind.Line, ChartKind.Area, ChartKind.StackedArea].indexOf(props.chartKind) >= 0
    })

    const showSymbol = computed(() => {
      return [ChartKind.Line, ChartKind.Area, ChartKind.StackedArea].indexOf(props.chartKind) >= 0
    })
    const symbolItems = computed(() => {
      const symbols = ['none', 'circle', 'rect', 'roundRect', 'triangle', 'diamond', 'pin', 'arrow']
      return symbols.map((symbol) => {
        return { value: symbol, text: symbol }
      })
    })

    return {
      dialog,
      showOpacity,
      showLineWidth,
      showSymbol,
      symbolItems,
    }
  },
})
</script>

<style lang="scss" scoped></style>
