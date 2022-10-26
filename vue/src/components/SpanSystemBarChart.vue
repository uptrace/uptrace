<template>
  <div>
    <v-sheet outlined rounded="t-lg">
      <v-row dense justify="space-around" class="pa-2">
        <v-col
          v-for="sys in styledSystems"
          :key="sys.system"
          justify="space-around"
          cols="auto"
          class="text-center text-subtitle-2"
        >
          <v-avatar :color="sys.color.base" size="12" class="mr-2"></v-avatar>
          <span class="d-inline-flex mr-1">{{ truncate(sys.system, { length: 32 }) }}</span>
          <span class="d-inline-flex blue-grey--text">{{
            percent(sys.duration / totalDuration)
          }}</span>
        </v-col>
      </v-row>
    </v-sheet>

    <div class="d-flex">
      <v-tooltip v-for="sys in styledSystems" :key="sys.system" bottom>
        <template #activator="{ on }">
          <div :style="sys.barStyle" class="bar" v-on="on"></div>
        </template>
        <div>
          <span>{{ sys.system }}</span>
          <XDuration :duration="sys.duration" class="ml-1" />
        </div>
      </v-tooltip>
    </div>

    <v-progress-linear v-show="loading" indeterminate absolute></v-progress-linear>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Utilities
import { ColoredSystem } from '@/models/colored-system'
import { percent } from '@/util/fmt'

export default defineComponent({
  name: 'SpanSystemBarChart',

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
    const totalDuration = computed((): number => {
      return props.systems.reduce((acc, system) => {
        if (system.duration) {
          return acc + system.duration
        }
        return acc
      }, 0)
    })

    const styledSystems = computed(() => {
      let systems = props.systems.slice(0, 5)

      for (let c of systems) {
        c.barStyle = {
          width: pct(c.duration, totalDuration.value),
          'background-color': c.color.base,
        }
      }

      return systems
    })

    return { styledSystems, totalDuration, truncate, percent }
  },
})

function pct(a: number, b: number) {
  if (b === 0) {
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
