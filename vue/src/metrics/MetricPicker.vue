<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-row align="start" class="mb-n5">
      <v-col cols="5">
        <v-autocomplete
          v-model="metricName"
          :loading="loading"
          :items="filteredMetrics"
          item-text="name"
          item-value="name"
          auto-select-first
          label="Select a metric..."
          :rules="rules.name"
          hide-details="auto"
          :disabled="disabled"
          solo
          flat
          dense
          background-color="grey lighten-4"
          :search-input.sync="searchInput"
          no-filter
          clearable
          @click:clear="reset"
          @change="onMetricNameChange"
        >
          <template #item="{ item }">
            <v-list-item-content>
              <v-list-item-title>
                <span>{{ item.name }}</span>
                <v-chip label small color="grey lighten-4" title="Instrument" class="ml-2">{{
                  item.instrument
                }}</v-chip>
                <v-chip
                  v-if="item.unit"
                  label
                  small
                  color="grey lighten-4"
                  title="Unit"
                  class="ml-2"
                  >{{ item.unit }}</v-chip
                >
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ item.description }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </template>
        </v-autocomplete>
      </v-col>
      <v-col cols="auto" class="mt-2 text--secondary">AS</v-col>
      <v-col cols="5" md="4">
        <v-text-field
          ref="metricAliasRef"
          v-model="metricAlias"
          label="Short alias"
          :rules="rules.alias"
          hide-details="auto"
          prefix="$"
          solo
          flat
          dense
          background-color="grey lighten-4"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-btn type="submit" color="primary" :disabled="!isValid">Add metric</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'
import { defaultMetricAlias } from '@/metrics/use-metrics'

// Utilities
import { unitShortName } from '@/util/fmt'
import { requiredRule, optionalRule } from '@/util/validation'

// Types
import { Metric, MetricAlias } from '@/metrics/types'

export default defineComponent({
  name: 'MetricPicker',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    activeMetrics: {
      type: Array as PropType<MetricAlias[]>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    required: {
      type: Boolean,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const metricAliasRef = shallowRef()
    const metricName = shallowRef('')
    const metricAlias = shallowRef('')
    const searchInput = shallowRef('')

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = computed(() => {
      return {
        name: [props.required ? requiredRule : optionalRule],
        alias: [
          (v: string) => {
            if (!metricName.value) {
              return true
            }
            if (!v) {
              return 'Alias is required'
            }
            if (!/^[a-z][a-z0-9_]*$/i.test(v)) {
              return 'Only letters and numbers are allowed'
            }

            const found = props.activeMetrics.find((m) => m.alias === v)
            if (found) {
              return 'Alias is duplicated'
            }

            return true
          },
        ],
      }
    })

    const filteredMetrics = computed((): Metric[] => {
      let metrics = props.metrics
      if (searchInput.value) {
        metrics = fuzzyFilter(metrics, searchInput.value, { key: 'name' })
      }
      return metrics
    })

    function submit() {
      if (!validate()) {
        return
      }

      metricAlias.value = metricAlias.value.toLowerCase()
      ctx.emit('click:add', { name: metricName.value, alias: metricAlias.value })
      reset()
    }

    function onMetricNameChange(metricName: string | null) {
      if (!metricName) {
        return
      }

      metricAlias.value = defaultMetricAlias(metricName)
      metricAliasRef.value.focus()
    }

    function validate() {
      return form.value.validate()
    }

    function reset() {
      metricName.value = ''
      metricAlias.value = ''
      form.value.reset()
    }

    return {
      searchInput,
      metricAliasRef,
      metricName,
      metricAlias,

      form,
      isValid,
      rules,
      filteredMetrics,

      submit,
      validate,
      reset,

      unitShortName,
      onMetricNameChange,
    }
  },
})
</script>

<style lang="scss" scoped></style>
