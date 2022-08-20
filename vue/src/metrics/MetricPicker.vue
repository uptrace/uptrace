<template>
  <v-form v-model="isValid">
    <v-row dense align="start">
      <v-col cols="auto">
        <v-btn icon @click="$emit('click:remove', value)"
          ><v-icon>mdi-minus-circle-outline</v-icon></v-btn
        >
      </v-col>
      <v-col cols="6" md="5">
        <v-autocomplete
          v-model="metricName"
          :items="filteredMetrics"
          item-text="name"
          item-value="name"
          auto-select-first
          label="Select a metric"
          :rules="rules.name"
          :disabled="disabled"
          dense
          solo
          flat
          background-color="grey lighten-4"
          hide-details="auto"
          :search-input.sync="searchInput"
          no-filter
          @change="onMetricNameChange"
        >
          <template #item="{ item }">
            <v-list-item-content>
              <v-list-item-title>
                <span>{{ item.name }}</span>
                <v-chip label small color="grey lighten-4" class="ml-2">{{
                  item.instrument
                }}</v-chip>
                <v-chip v-if="item.unit" label small color="grey lighten-4" class="ml-2">{{
                  item.unit
                }}</v-chip>
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ item.description }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </template>
        </v-autocomplete>
      </v-col>
      <v-col cols="auto">
        <div class="mt-2 mx-2 text--disabled">AS</div>
      </v-col>
      <v-col cols="4" md="3">
        <v-text-field
          ref="metricAliasRef"
          v-model="metricAlias"
          label="Short alias"
          :rules="rules.alias"
          prefix="$"
          solo
          flat
          dense
          background-color="grey lighten-4"
          hide-details="auto"
        />
      </v-col>
      <v-col cols="auto">
        <v-btn dense solo :disabled="applyDisabled" class="ml-2" @click="apply">Apply</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'
import { hasMetricAlias } from '@/metrics/use-query'

// Utilities
import { unitShortName } from '@/util/fmt'
import { requiredRule } from '@/util/validation'

// Types
import { Metric, MetricAlias, Instrument } from '@/metrics/types'

export default defineComponent({
  name: 'MetricPicker',

  props: {
    value: {
      type: Object as PropType<MetricAlias>,
      required: true,
    },
    index: {
      type: Number,
      required: true,
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
    disabled: {
      type: Boolean,
      default: false,
    },
    required: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const metricAliasRef = shallowRef()
    const metricName = shallowRef('')
    const metricAlias = shallowRef('')
    const searchInput = shallowRef('')

    const isValid = shallowRef(false)
    const rules = computed(() => {
      return {
        name: props.required ? [requiredRule] : [],
        alias: [
          (v: string) => {
            if (!props.required && !metricName.value) {
              return true
            }
            if (!v) {
              return 'Alias is required'
            }
            if (!/^[a-z]/i.test(v)) {
              return 'Must start with a letter'
            }
            if (!/^[a-z][a-z0-9_]*$/i.test(v)) {
              return 'Only letters and numbers are allowed'
            }

            const activeIndex = props.activeMetrics.findIndex((m) => m.alias === v)
            if (activeIndex >= 0 && activeIndex !== props.index) {
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

      if (props.value.name) {
        const i = metrics.findIndex((m) => m.name === props.value.name)
        if (i >= 0) {
          return metrics
        }

        let found = props.metrics.find((m) => m.name === props.value.name)
        if (!found) {
          found = { name: props.value.name, instrument: Instrument.Invalid } as Metric
        }
        metrics.push(found)
      }

      return metrics
    })

    const applyDisabled = computed((): boolean => {
      if (!metricName.value || !metricAlias.value) {
        return true
      }
      if (metricName.value !== props.value.name || metricAlias.value !== props.value.alias) {
        return false
      }
      if (!hasMetricAlias(props.uql.query, metricAlias.value)) {
        return false
      }
      return true
    })

    watch(
      () => props.value,
      (metric: MetricAlias) => {
        metricName.value = metric.name
        metricAlias.value = metric.alias
      },
      { immediate: true },
    )

    function apply() {
      metricAlias.value = metricAlias.value.toLowerCase()
      ctx.emit('click:apply', { name: metricName.value, alias: metricAlias.value })
    }

    function onMetricNameChange(metricName: string | null) {
      if (!metricName) {
        return
      }

      let alias = metricName

      const i = alias.lastIndexOf('.')
      if (i >= 0) {
        alias = alias.slice(i + 1)
      }

      metricAlias.value = alias
      metricAliasRef.value.focus()
    }

    return {
      searchInput,
      metricAliasRef,
      metricName,
      metricAlias,

      isValid,
      rules,
      filteredMetrics,
      applyDisabled,

      unitShortName,
      apply,
      onMetricNameChange,
    }
  },
})
</script>

<style lang="scss" scoped></style>
