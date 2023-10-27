<template>
  <v-form ref="formRef" v-model="isValid" @submit.prevent="apply">
    <v-row
      :dense="$vuetify.breakpoint.mdAndDown"
      align="start"
      :class="$vuetify.breakpoint.mdAndDown ? 'mb-n4' : 'mb-n8'"
    >
      <v-col v-if="editable" cols="auto">
        <v-btn
          icon
          title="Remove metric"
          :disabled="!('click:remove' in $listeners)"
          @click="$emit('click:remove')"
        >
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </v-col>
      <v-col cols="5" md="4">
        <v-autocomplete
          v-model="metricName"
          :loading="loading"
          :items="filteredMetrics"
          item-text="name"
          item-value="name"
          auto-select-first
          label="Select a metric..."
          :rules="rules.name"
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
      <v-col cols="4" md="3" lg="2">
        <v-text-field
          ref="metricAliasRef"
          v-model="metricAlias"
          placeholder="short_alias"
          :rules="rules.alias"
          prefix="$"
          solo
          flat
          dense
          clearable
          background-color="grey lighten-4"
        />
      </v-col>
      <v-col cols="auto">
        <v-btn type="submit" color="primary" :disabled="applyDisabled">Apply</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { defaultMetricAlias } from '@/metrics/use-metrics'

// Utilities
import { unitShortName } from '@/util/fmt'
import { requiredRule } from '@/util/validation'
import { escapeRe } from '@/util/string'

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

    value: {
      type: Object as PropType<MetricAlias>,
      default: undefined,
    },
    activeMetrics: {
      type: Array as PropType<MetricAlias[]>,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const metricAliasRef = shallowRef()
    const searchInput = shallowRef('')

    const metricName = shallowRef(props.value?.name ?? '')
    const metricAlias = shallowRef(props.value?.alias ?? '')

    const formRef = shallowRef()
    const isValid = shallowRef(false)
    const rules = computed(() => {
      return {
        name: [requiredRule],
        alias: [
          (v: string) => {
            if (!metricName.value) {
              return true
            }
            if (!v) {
              return 'Alias is required'
            }
            if (!/^[a-z0-9_]*$/i.test(v)) {
              return 'Only letters and numbers are allowed'
            }

            if (v !== props.value?.alias) {
              const found = props.activeMetrics.find((m) => m.alias === v)
              if (found) {
                return 'Alias is duplicated'
              }
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

    const applyDisabled = computed((): boolean => {
      if (!metricName.value || !metricAlias.value) {
        return true
      }
      if (props.value) {
        if (metricName.value !== props.value.name || metricAlias.value !== props.value.alias) {
          return false
        }
      }
      if (!createRegexp(metricAlias.value).test(props.query)) {
        return false
      }
      return true
    })

    function apply() {
      metricAlias.value = metricAlias.value.toLowerCase()
      ctx.emit('click:apply', { name: metricName.value, alias: metricAlias.value })
      if (!props.value) {
        reset()
      }
    }

    function reset() {
      metricName.value = ''
      metricAlias.value = ''
      formRef.value.reset()
    }

    function onMetricNameChange(metricName: string | null) {
      if (!metricName) {
        return
      }

      metricAlias.value = defaultMetricAlias(metricName)
      metricAliasRef.value.focus()
    }

    return {
      searchInput,
      metricAliasRef,
      metricName,
      metricAlias,

      formRef,
      isValid,
      rules,
      filteredMetrics,

      applyDisabled,
      apply,
      reset,

      unitShortName,
      onMetricNameChange,
    }
  },
})

function createRegexp(alias: string, flags = '') {
  return new RegExp(escapeRe('$' + alias) + '\\b', flags)
}
</script>

<style lang="scss" scoped></style>
