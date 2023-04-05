<template>
  <v-card flat :max-width="width">
    <v-container fluid>
      <v-row align="center">
        <v-col cols="3" class="text--secondary">Color</v-col>
        <v-col>
          <v-btn outlined @click="dialog = true">
            <v-icon size="24" :color="timeseriesStyle.color" left>mdi-circle</v-icon>
            <span>{{ timeseriesStyle.color }}</span>
          </v-btn>
        </v-col>
      </v-row>

      <v-row v-if="showOpacity" align="center">
        <v-col cols="3" class="text--secondary">Fill opacity</v-col>
        <v-col>
          <v-slider v-model="timeseriesStyle.opacity" min="0" max="100" hide-details="auto">
            <template #prepend>{{ timeseriesStyle.opacity }}%</template>
          </v-slider>
        </v-col>
      </v-row>

      <v-row v-if="showLineWidth" align="center">
        <v-col cols="3" class="text--secondary">Line width</v-col>
        <v-col>
          <v-slider v-model="timeseriesStyle.lineWidth" min="1" max="10" hide-details="auto">
            <template #prepend>{{ timeseriesStyle.lineWidth }}px</template>
          </v-slider>
        </v-col>
      </v-row>

      <v-row v-if="showSymbol" align="center">
        <v-col cols="3" class="text--secondary">Symbol</v-col>
        <v-col>
          <v-select
            v-model="timeseriesStyle.symbol"
            :items="symbolItems"
            filled
            dense
            hide-details="auto"
          ></v-select>
        </v-col>
      </v-row>

      <v-row v-if="showSymbol" align="center">
        <v-col cols="3" class="text--secondary">Symbol size</v-col>
        <v-col>
          <v-slider v-model="timeseriesStyle.symbolSize" min="1" max="16" hide-details="auto">
            <template #prepend>{{ timeseriesStyle.symbolSize }}px</template>
          </v-slider>
        </v-col>
      </v-row>

      <v-row>
        <v-spacer />
        <v-col cols="auto">
          <v-btn text @click="$emit('click:reset')">Reset to defaults</v-btn>
          <v-btn color="primary" @click="$emit('click:ok')">OK</v-btn>
        </v-col>
      </v-row>

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
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Types
import { ChartKind, TimeseriesStyle } from '@/metrics/types'

export default defineComponent({
  name: 'TimeseriesStyleCard',

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
