<template>
  <v-form v-model="isValid" @submit.prevent="update">
    <v-card>
      <v-toolbar color="light-blue lighten-5" flat dense>
        <v-toolbar-title>Edit dashboard</v-toolbar-title>
      </v-toolbar>

      <div class="py-4 px-6">
        <v-row class="mb-n2">
          <v-col>
            <v-text-field
              v-model="dashboard.name"
              label="Dashboard name"
              :rules="rules.name"
              dense
              filled
              background-color="grey lighten-4"
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
              hint="Shift time to the past (negative offset) or future (positive offset)"
              placeholder="-60"
              suffix="minutes"
              persistent-hint
              filled
              dense
              :rules="rules.timeOffset"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <v-btn color="primary" text @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn type="submit" color="primary" :disabled="!isValid" :loading="dashMan.pending"
              >Update</v-btn
            >
          </v-col>
        </v-row>
      </div>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useDashManager } from '@/metrics/use-dashboards'

// Utilities
import { Dashboard } from '@/metrics/types'
import { requiredRule, minMaxRule } from '@/util/validation'
import { MINUTE } from '@/util/fmt/date'

export default defineComponent({
  name: 'DashEditForm',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const name = shallowRef('')
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

    const dashMan = useDashManager()

    function update() {
      if (!isValid.value) {
        return
      }

      dashMan
        .update({
          name: props.dashboard.name,
          minInterval: props.dashboard.minInterval,
          timeOffset: props.dashboard.timeOffset,
        })
        .then((dash) => {
          ctx.emit('update', dash)
        })
    }

    return {
      name,
      isValid,
      rules,
      minIntervalItems,
      timeOffset,

      dashMan,
      update,
    }
  },
})
</script>

<style lang="scss" scoped></style>
