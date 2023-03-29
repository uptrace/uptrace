<template>
  <v-row dense justify="center">
    <DashGaugeRowCol
      v-for="gauge in dashGauges"
      :key="gauge.id"
      :date-range="dateRange"
      :dash-gauge="gauge"
      :grid-query="gridQuery"
      :editable="editable"
      @click:edit="openDialog($event)"
      @change="$emit('change', $event)"
    />
    <v-col cols="auto">
      <v-card
        width="200"
        height="100%"
        min-height="92"
        outlined
        rounded="lg"
        class="d-flex align-center"
      >
        <v-card-text class="text-center">
          <v-btn icon x-large @click="openDialog()">
            <v-icon size="50" color="grey lighten-2">mdi-plus</v-icon>
          </v-btn>
        </v-card-text>
      </v-card>
    </v-col>

    <v-dialog v-model="dialog" max-width="1200">
      <DashGaugeForm
        v-if="dialog && activeDashGauge"
        :date-range="dateRange"
        :dash-gauge="activeDashGauge"
        :editable="editable"
        @click:save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
    </v-dialog>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { emptyDashGauge } from '@/metrics/gauge/use-dash-gauges'

// Components
import DashGaugeRowCol from '@/metrics/gauge/DashGaugeRowCol.vue'
import DashGaugeForm from '@/metrics/gauge/DashGaugeForm.vue'

// Utilities
import { DashKind, DashGauge } from '@/metrics/types'

export default defineComponent({
  name: 'DashGaugeRow',
  components: { DashGaugeRowCol, DashGaugeForm },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashKind: {
      type: String as PropType<DashKind>,
      required: true,
    },
    gridQuery: {
      type: String,
      default: '',
    },
    editable: {
      type: Boolean,
      default: false,
    },
    dashGauges: {
      type: Array as PropType<DashGauge[]>,
      required: true,
    },
  },

  setup(props) {
    const internalDialog = shallowRef(false)
    const activeDashGauge = shallowRef<DashGauge>()

    const dialog = computed({
      get(): boolean {
        return Boolean(internalDialog.value && activeDashGauge.value)
      },
      set(dialog: boolean) {
        internalDialog.value = dialog
      },
    })

    function openDialog(dashGauge: DashGauge | undefined) {
      if (!dashGauge) {
        dashGauge = emptyDashGauge(props.dashKind)
      }
      activeDashGauge.value = reactive(dashGauge)
      dialog.value = true
    }

    return {
      dialog,
      openDialog,

      activeDashGauge,
    }
  },
})
</script>

<style lang="scss" scoped></style>
