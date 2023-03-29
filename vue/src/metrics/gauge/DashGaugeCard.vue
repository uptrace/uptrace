<template>
  <v-card
    :loading="loading"
    min-width="100"
    outlined
    rounded="lg"
    class="border-bottom"
    :style="style"
  >
    <div class="py-2 px-3">
      <v-row align="center" dense>
        <v-col cols="auto">
          <v-tooltip top>
            <template #activator="{ on, attrs }">
              <div
                class="text-no-wrap body-2 blue-grey--text text--lighten-1"
                v-bind="attrs"
                v-on="on"
              >
                {{ dashGauge.name }}
              </div>
            </template>
            <span>{{ dashGauge.description || dashGauge.name }}</span>
          </v-tooltip>
        </v-col>
        <v-col v-if="showEdit" cols="auto">
          <v-menu offset-y>
            <template #activator="{ on, attrs }">
              <v-btn :loading="dashGaugeMan.pending" icon v-bind="attrs" v-on="on">
                <v-icon>mdi-dots-vertical</v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item @click="$emit('click:edit', dashGauge)">
                <v-list-item-icon>
                  <v-icon>{{ editable ? 'mdi-pencil' : 'mdi-lock' }}</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>Edit</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-if="editable" @click="del">
                <v-list-item-icon>
                  <v-icon>mdi-delete</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>Delete</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-col>
      </v-row>

      <v-row dense>
        <v-col class="py-2 text-h5 text-center">{{ gaugeText }}</v-col>
      </v-row>
    </div>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'
import colors from 'vuetify/lib/util/colors'

// Composables
import { formatGauge, useDashGaugeManager } from '@/metrics/gauge/use-dash-gauges'

// Utilities
import { DashGauge, StyledColumnInfo } from '@/metrics/types'

export default defineComponent({
  name: 'DashGaugeCard',

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    dashGauge: {
      type: Object as PropType<DashGauge>,
      required: true,
    },
    columns: {
      type: Array as PropType<StyledColumnInfo[]>,
      required: true,
    },
    values: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    showEdit: {
      type: Boolean,
      default: false,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const gaugeText = computed(() => {
      return formatGauge(props.values, props.columns, props.dashGauge.template)
    })

    const color = computed(() => {
      for (let col of props.columns) {
        if (col.color) {
          return col.color
        }
      }
      return colors.blue.darken2
    })

    const style = computed(() => {
      return {
        'border-bottom-color': color.value,
      }
    })

    const dashGaugeMan = useDashGaugeManager()

    function del() {
      dashGaugeMan.del(props.dashGauge).then(() => {
        ctx.emit('change')
      })
    }

    return {
      gaugeText,
      style,

      dashGaugeMan,
      del,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom-width: 8px;
}
</style>
