<template>
  <v-form v-model="isValid" @submit.prevent="submit">
    <v-card>
      <v-toolbar color="bg--none-primary" flat>
        <v-toolbar-title>{{ dashboard.id ? 'Edit' : 'New' }} dashboard</v-toolbar-title>

        <v-spacer />

        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid>
        <v-row dense>
          <v-col>
            <v-text-field
              v-model="dashboard.name"
              label="Dashboard name"
              :rules="rules.name"
              dense
              filled
              required
              autofocus
            />
          </v-col>
        </v-row>

        <v-row dense>
          <v-col>
            <v-select
              v-model="dashboard.minInterval"
              :items="minIntervalItems"
              label="Min interval"
              hint="Min limit for the automatically calculated interval"
              persistent-hint
              dense
              filled
            ></v-select>
          </v-col>
        </v-row>

        <v-row dense>
          <v-col>
            <v-text-field
              v-model.number="timeOffset"
              type="number"
              label="Time offset"
              hint="Shift time to the past (positive offset) or future (negative offset)"
              placeholder="15"
              suffix="minutes"
              persistent-hint
              filled
              dense
              :rules="rules.timeOffset"
            />
          </v-col>
        </v-row>

        <v-row dense>
          <v-col>
            <v-select
              v-model="dashboard.gridMaxWidth"
              :items="gridMaxWidthItems"
              label="Grid max width"
              filled
              dense
            />
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <v-btn color="primary" text @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn type="submit" color="primary" :disabled="!isValid" :loading="dashMan.pending">
              {{ dashboard.id ? 'Update' : 'Create' }}
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, computed, PropType } from 'vue'

// Composables
import { useDashboardManager } from '@/metrics/use-dashboards'

// Misc
import { Dashboard } from '@/metrics/types'
import { requiredRule, minMaxRule } from '@/util/validation'
import { MINUTE } from '@/util/fmt/date'

export default defineComponent({
  name: 'DashboardForm',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      default: () => emptyDashboard(),
    },
  },

  setup(props, ctx) {
    const isValid = shallowRef(false)
    const rules = {
      name: [requiredRule],
      timeOffset: [minMaxRule(-300, 300)],
    }

    const minIntervalItems = computed(() => {
      return [
        { text: 'Not set', value: 0 },
        { text: '2 minutes', value: 2 * MINUTE },
        { text: '3 minutes', value: 3 * MINUTE },
        { text: '5 minutes', value: 5 * MINUTE },
        { text: '10 minutes', value: 10 * MINUTE },
        { text: '15 minutes', value: 15 * MINUTE },
      ]
    })

    const timeOffset = computed({
      get() {
        return props.dashboard.timeOffset / MINUTE
      },
      set(minutes: number) {
        props.dashboard.timeOffset = minutes * MINUTE
      },
    })

    const gridMaxWidthItems = computed(() => {
      return [
        { text: '1400px', value: 1416 },
        { text: '1600px', value: 1600 },
        { text: '1800px', value: 1800 },
        { text: '2000px', value: 2000 },
      ]
    })

    const dashMan = useDashboardManager()

    function submit() {
      if (!isValid.value) {
        return
      }

      if (props.dashboard.id) {
        dashMan.update(props.dashboard).then((dash) => {
          ctx.emit('saved', dash)
        })
      } else {
        dashMan.create(props.dashboard).then((dash) => {
          ctx.emit('saved', dash)
        })
      }
    }

    return {
      isValid,
      rules,
      minIntervalItems,
      timeOffset,
      gridMaxWidthItems,

      dashMan,
      submit,

      emptyDashboard,
    }
  },
})

function emptyDashboard(): Dashboard {
  return reactive({
    id: 0,
    projectId: 0,
    templateId: '',

    name: '',
    pinned: false,

    minInterval: 0,
    timeOffset: 0,

    tableMetrics: [],
    tableQuery: '',
    tableGrouping: [],
    tableColumnMap: {},

    gridQuery: '',
    gridMaxWidth: 1416,
  })
}
</script>

<style lang="scss" scoped></style>
