<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-card>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>
          {{ dashGauge.id ? 'Edit text gauge' : 'New text gauge' }}
        </v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/get/querying-metrics.html" target="_blank"
          ><v-icon>mdi-help-circle-outline</v-icon></v-btn
        >

        <v-spacer />

        <v-btn
          small
          outlined
          :loading="gaugeQuery.loading"
          class="mr-4"
          @click="gaugeQuery.reload()"
        >
          <v-icon small left>mdi-refresh</v-icon>
          <span>Reload</span>
        </v-btn>

        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid class="pa-6">
        <v-row>
          <v-col class="text-subtitle-1">
            Text gauges are like <code>sprintf(format, values)</code>. You specify a
            <code>format</code> string with substitutions and Uptrace provides values.<br />
            For example, using <code>${up_dbs} out of ${total_dbs} are up</code> format string you
            will get <code>5 out of 5 dbs are up</code> as the result.
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select metrics that you want to use as values in the format string.
            </v-sheet>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <MetricsPicker
              ref="metricsPicker"
              v-model="dashGauge.metrics"
              :uql="uql"
              :editable="editable"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Add some aggregations and filters, for example,
              <code>uniq($metric_name.host.name) as num_host</code>.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              :disabled="!activeMetrics.length"
              show-agg
              show-dash-where
            />
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col cols="3" class="text--secondary">Formatted value</v-col>
          <v-col cols="auto" class="font-weight-medium">
            {{ gaugeText }}
          </v-col>
          <v-col v-for="(col, colName) in dashGauge.columnMap" :key="colName" cols="auto">
            <MetricColumnChip :name="colName" :column="col" />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="mt-3 text--secondary">Optional format</v-col>
          <v-col cols="9">
            <v-text-field
              v-model="dashGauge.template"
              placeholder="${sum($db_up)} dbs up out of ${count($db_up)}"
              hint="Format string to customize the gauge"
              persistent-hint
              filled
              dense
              clearable
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Gauge name</v-col>
          <v-col cols="9">
            <v-text-field
              v-model="dashGauge.name"
              label="Name"
              hint="Short name that describes the gauge"
              persistent-hint
              filled
              dense
              :rules="rules.name"
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Gauge description</v-col>
          <v-col cols="9">
            <v-text-field
              v-model="dashGauge.description"
              label="Description"
              hint="Description or comment"
              persistent-hint
              filled
              dense
              :rules="rules.description"
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="text--secondary">Preview</v-col>
          <v-col cols="auto">
            <DashGaugeCard
              :loading="gaugeQuery.loading"
              :dash-gauge="dashGauge"
              :columns="gaugeQuery.columns"
              :values="gaugeQuery.values"
              :column-map="dashGauge.columnMap"
            />
          </v-col>
        </v-row>

        <v-row v-if="editable" class="mt-8">
          <v-spacer />
          <v-col cols="auto">
            <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn
              type="submit"
              color="primary"
              :disabled="!isValid"
              :loading="dashGaugeMan.pending"
              >{{ dashGauge.id ? 'Save' : 'Create' }}</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import {
  formatGauge,
  useDashGaugeManager,
  useDashGaugeQuery,
} from '@/metrics/gauge/use-dash-gauges'

// Components
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import MetricColumnChip from '@/metrics/MetricColumnChip.vue'

// Utilities
import { requiredRule } from '@/util/validation'
import { updateColumnMap, DashGauge } from '@/metrics/types'

export default defineComponent({
  name: 'DashGaugeForm',
  components: {
    MetricsPicker,
    MetricsQueryBuilder,
    MetricColumnChip,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashGauge: {
      type: Object as PropType<DashGauge>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const metricsPicker = shallowRef()
    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = { metrics: [requiredRule], name: [requiredRule], description: [requiredRule] }

    const uql = useUql()
    const dashGaugeMan = useDashGaugeManager()

    const activeMetrics = useActiveMetrics(computed(() => props.dashGauge.metrics))

    const gaugeQuery = useDashGaugeQuery(
      () => {
        if (!props.dashGauge || !props.dashGauge.metrics.length || !props.dashGauge.query) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.dashGauge.metrics.map((m) => m.name),
          alias: props.dashGauge.metrics.map((m) => m.alias),
          query: props.dashGauge.query,
        }
      },
      computed(() => props.dashGauge.columnMap),
    )

    const gaugeText = computed(() => {
      return formatGauge(
        gaugeQuery.values,
        gaugeQuery.columns,
        props.dashGauge.template,
        'Select a metric first...',
      )
    })

    watch(
      () => props.dashGauge.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        props.dashGauge.query = query
      },
    )

    watch(
      () => gaugeQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
    )

    watch(
      () => gaugeQuery.columns,
      (columns) => {
        updateColumnMap(props.dashGauge.columnMap, columns)
      },
    )

    function submit() {
      const r1 = metricsPicker.value.validate()
      const r2 = form.value.validate()
      if (!r1 || !r2) {
        return
      }

      dashGaugeMan.save(props.dashGauge).then((dashGauge) => {
        ctx.emit('click:save', dashGauge)
      })
    }

    return {
      uql,
      dashGaugeMan,

      activeMetrics,
      gaugeQuery,
      gaugeText,

      metricsPicker,
      form,
      isValid,
      rules,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
