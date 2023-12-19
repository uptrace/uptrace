<template>
  <div>
    <v-sheet outlined rounded="t-lg">
      <v-row dense justify="space-around" class="pa-2">
        <v-col
          v-for="sys in styledSystems"
          :key="sys.system"
          cols="auto"
          class="text-center text-subtitle-2"
        >
          <v-avatar :color="sys.color" size="12" class="mr-2"></v-avatar>
          <span class="d-inline-flex mr-1">{{ truncate(sys.name, { length: 32 }) }}</span>
          <PctValue
            :a="sys.duration"
            :b="totalDuration"
            :unit="Unit.Microseconds"
            class="d-inline-flex blue-grey--text"
          />
        </v-col>
      </v-row>
    </v-sheet>

    <div class="d-flex">
      <v-tooltip v-for="sys in styledSystems" :key="sys.name" bottom>
        <template #activator="{ on }">
          <div :style="sys.barStyle" class="bar" v-on="on"></div>
        </template>
        <div>
          <span>{{ sys.name }}</span>
          <DurationValue :value="sys.duration" class="ml-1" />
        </div>
      </v-tooltip>
    </div>

    <v-progress-linear v-show="loading" indeterminate absolute></v-progress-linear>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Components
import PctValue from '@/components/PctValue.vue'

// Misc
import { Unit } from '@/util/fmt'
import { ColoredSystem } from '@/models/colored-system'

export default defineComponent({
  name: 'SpanSystemBarChart',
  components: { PctValue },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    systems: {
      type: Array as PropType<ColoredSystem[]>,
      required: true,
    },
  },

  setup(props) {
    const totalDuration = computed(() => {
      return props.systems.reduce((acc, system) => {
        return acc + system.duration
      }, 0)
    })

    const styledSystems = computed(() => {
      return props.systems.map((system) => {
        return {
          ...system,
          barStyle: {
            width: pct(system.duration, totalDuration.value),
            'background-color': system.color,
          },
        }
      })
    })

    return {
      Unit,

      totalDuration,
      styledSystems,

      truncate,
    }
  },
})

function pct(a: number, b: number) {
  if (b === 0 || a >= b) {
    return '100%'
  }
  const pct = (a / b) * 100
  return pct + '%'
}
</script>

<style lang="scss" scoped>
.bar {
  height: 14px;
}
</style>
