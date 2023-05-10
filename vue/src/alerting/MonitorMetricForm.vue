<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-row align="center">
      <v-col cols="auto" class="pr-4">
        <v-avatar color="blue darken-1" size="40">
          <span class="white--text text-h5">1</span>
        </v-avatar>
      </v-col>
      <v-col class="text-h5">Select metrics to monitor</v-col>
    </v-row>

    <v-row>
      <v-col class="text-subtitle-1 text--primary">
        You can monitor any query that returns a single column.<br />
        To monitor multiple
        <a href="https://uptrace.dev/opentelemetry/metrics.html#timeseries" target="_blank"
          >timeseries</a
        >
        at once, add <code>group by</code> clause, for example, <code>group by host.name</code>.
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <MetricsPicker
          v-model="monitor.params.metrics"
          :metrics="metrics.items"
          :uql="uql"
          editable
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <MetricsQueryBuilder
          :date-range="dateRange"
          :metrics="activeMetrics"
          :uql="uql"
          :disabled="!activeMetrics.length"
          show-group-by
          show-metrics-where
        />

        <div v-if="Object.keys(columnMap).length > 1" class="mt-1 d-flex align-center">
          <div>
            <v-icon size="30" color="red darken-1" class="mr-3">mdi-alert-circle</v-icon>
          </div>
          <div class="text-body-2">
            The query returns {{ Object.keys(columnMap).length }} columns, but only a single column
            is allowed.<br />
            To keep the column but hide the result, underscore the alias, for example,
            <code>count($metric) as _tmp_count</code>.
          </div>
        </div>
        <div
          v-else-if="timeseries.status.hasData() && Object.keys(columnMap).length === 0"
          class="mt-1 d-flex align-center"
        >
          <div>
            <v-icon size="30" color="red darken-1" class="mr-3">mdi-alert-circle</v-icon>
          </div>
          <div class="text-body-2">The query must return at least one column to monitor.</div>
        </div>
      </v-col>
    </v-row>

    <template v-if="timeseries.status.hasData()">
      <v-row>
        <v-col>
          <v-chip v-for="(col, colName) in columnMap" :key="colName" outlined label class="ma-1">
            <span>{{ colName }}</span>
            <UnitPicker v-model="col.unit" target-class="mr-n4" />
          </v-chip>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <MetricChart
            :loading="timeseries.loading"
            :resolved="timeseries.status.isResolved()"
            :timeseries="styledTimeseries"
            :time="timeseries.time"
            :min-allowed-value="monitor.params.minValue"
            :max-allowed-value="monitor.params.maxValue"
            :event-bus="eventBus"
          />
        </v-col>
      </v-row>
      <v-row v-if="timeseries.items.length" no-gutters justify="center">
        <v-col cols="auto">
          <ChartLegendTable
            :timeseries="styledTimeseries"
            @hover:item="eventBus.emit('hover', $event)"
          />
        </v-col>
      </v-row>
    </template>

    <v-row>
      <v-col>
        <v-divider />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="auto" class="pr-4">
        <v-avatar color="blue darken-1" size="40">
          <span class="white--text text-h5">2</span>
        </v-avatar>
      </v-col>
      <v-col cols="auto" class="text-h5">Specify alert trigger conditions</v-col>
    </v-row>

    <v-row class="mb-n6">
      <v-col>
        <p class="text--secondary">
          Specify the range of allowed values. The range will be highlighted in green on the chart
          above. Uptrace will create
          <router-link :to="{ name: 'AlertList' }">alerts</router-link> for values outside of the
          allowed range.
        </p>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-3 text--secondary">Allowed values range</v-col>
      <v-col cols="4">
        <v-text-field
          v-model.number="monitor.params.minValue"
          type="number"
          label="Min allowed value"
          :suffix="activeColumn?.unit"
          hint="Leave empty to disable"
          persistent-hint
          filled
          dense
          clearable
          :rules="rules.minValue"
          hide-details="auto"
        />
      </v-col>
      <v-col v-if="observedMin" cols="auto" class="mt-4 text-body-2 text--secondary">
        Observed min:
        <strong>{{ formatNum(observedMin) }}</strong>
        (zeroes excluded)
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3"></v-col>
      <v-col cols="4">
        <v-text-field
          v-model.number="monitor.params.maxValue"
          type="number"
          label="Max allowed value"
          :suffix="activeColumn?.unit"
          hint="Leave empty to disable"
          persistent-hint
          filled
          dense
          clearable
          :rules="rules.maxValue"
          hide-details="auto"
        />
      </v-col>
      <v-col v-if="observedMax" cols="auto" class="mt-4 text-body-2 text--secondary">
        Observed max:
        <strong>{{ formatNum(observedMax) }}</strong>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-4 text--secondary">Minimal duration</v-col>
      <v-col cols="6">
        <v-select
          v-model="monitor.params.forDuration"
          hint="Trigger an alert after this number of minutes"
          persistent-hint
          :items="forMinuteItems"
          filled
          dense
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-divider />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="auto" class="pr-4">
        <v-avatar color="blue darken-1" size="40">
          <span class="white--text text-h5">3</span>
        </v-avatar>
      </v-col>
      <v-col class="text-h5">Select notification channels</v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3" class="text--secondary">Email notifications</v-col>
      <v-col class="d-flex align-center">
        <v-checkbox
          v-model="monitor.notifyEveryoneByEmail"
          label="Notify everyone by email"
          hide-details="auto"
          class="mt-0"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-3 text--secondary">Slack and PagerDuty</v-col>
      <v-col cols="9" md="6">
        <v-select
          v-model="monitor.channelIds"
          multiple
          label="Notification channels"
          filled
          dense
          :items="channels.items"
          item-text="name"
          item-value="id"
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
      <v-col cols="3" class="mt-4 text--secondary">Monitor name</v-col>
      <v-col cols="9">
        <v-text-field
          v-model="monitor.name"
          label="Name"
          hint="Short name that describes the monitor"
          persistent-hint
          filled
          :rules="rules.name"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-spacer />
      <v-col cols="auto" class="pa-6">
        <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
        <v-btn type="submit" color="primary" :disabled="!isValid" :loading="monitorMan.pending">{{
          monitor.id ? 'Save' : 'Create'
        }}</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import numbro from 'numbro'
import { defineComponent, shallowRef, reactive, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useNotifChannels } from '@/alerting/use-notif-channels'
import { useActiveMetrics, UseMetrics } from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'
import { useMonitorManager, MetricMonitor } from '@/alerting/use-monitors'

// Components
import UnitPicker from '@/components/UnitPicker.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import MetricChart from '@/metrics/MetricChart.vue'
import ChartLegendTable from '@/metrics/ChartLegendTable.vue'

// Utilities
import { EventBus } from '@/models/eventbus'
import { updateColumnMap, MetricColumn } from '@/metrics/types'
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'MonitorMetricForm',
  components: {
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    MetricChart,
    ChartLegendTable,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
    monitor: {
      type: Object as PropType<MetricMonitor>,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      default: () => {
        return reactive({})
      },
    },
  },

  setup(props, ctx) {
    const eventBus = new EventBus()
    const channels = useNotifChannels(() => {
      return {}
    })

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = {
      name: [requiredRule],
      minValue: [
        (v: any) => {
          if (
            typeof props.monitor.params.minValue !== 'number' &&
            typeof props.monitor.params.maxValue !== 'number'
          ) {
            return 'At least min value is required'
          }
          return true
        },
      ],
      maxValue: [
        (v: any) => {
          if (
            typeof props.monitor.params.minValue !== 'number' ||
            typeof props.monitor.params.maxValue !== 'number'
          ) {
            return true
          }
          if (props.monitor.params.maxValue <= props.monitor.params.minValue) {
            return 'Max value should be greater than min'
          }
          return true
        },
      ],
    }
    const forMinuteItems = [
      { text: '1 minute', value: 1 },
      { text: '3 minutes', value: 3 },
      { text: '5 minutes', value: 5 },
      { text: '10 minutes', value: 10 },
      { text: '15 minutes', value: 15 },
    ]

    const uql = useUql()
    const monitorMan = useMonitorManager()
    const activeMetrics = useActiveMetrics(computed(() => props.monitor.params.metrics))
    const axiosParams = computed(() => {
      if (!props.monitor.params.query) {
        return undefined
      }

      const metrics = props.monitor.params.metrics.filter((m) => m.name && m.alias)
      if (!metrics.length) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: metrics.map((m) => m.name),
        alias: metrics.map((m) => m.alias),
        query: props.monitor.params.query,
      }
    })

    const timeseries = useTimeseries(() => {
      return axiosParams.value
    })

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => props.columnMap),
      computed(() => ({})),
    )

    const activeColumn = computed(() => {
      const columns = Object.keys(props.columnMap)

      if (columns.length !== 1) {
        return undefined
      }

      const colName = columns[0]
      const col = props.columnMap[colName]
      return {
        ...col,
        name: colName,
      }
    })

    const observedMin = computed(() => {
      let min = Number.MAX_VALUE
      for (let ts of timeseries.items) {
        for (let num of ts.value) {
          if (num === 0) {
            continue
          }
          if (num < min) {
            min = num
          }
        }
      }
      if (min !== Number.MAX_VALUE) {
        return min
      }
      return undefined
    })

    const observedMax = computed(() => {
      let max = 0
      for (let ts of timeseries.items) {
        if (ts.max > max) {
          max = ts.max
        }
      }
      return max
    })

    const observedAvg = computed(() => {
      let sum = 0
      let count = 0
      for (let ts of timeseries.items) {
        for (let num of ts.value) {
          sum += num
          count++
        }
      }
      return sum / count
    })

    watch(
      () => props.monitor.params.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        props.monitor.params.query = query
      },
    )

    watch(
      () => timeseries.query,
      (queryInfo) => {
        if (queryInfo) {
          uql.setQueryInfo(queryInfo)
        }
      },
      { immediate: true },
    )

    watch(
      () => timeseries.columns,
      (columns) => {
        updateColumnMap(props.columnMap, columns)

        const params = props.monitor.params
        if (params.column && params.columnUnit) {
          props.columnMap[params.column] = {
            unit: params.columnUnit,
            color: '',
          }
        }
      },
    )

    watch(
      () => props.monitor.params.minValue,
      () => form.value.validate(),
    )
    watch(
      () => props.monitor.params.maxValue,
      () => form.value.validate(),
    )

    function submit() {
      save().then(() => {
        ctx.emit('click:save')
      })
    }

    function save() {
      if (!form.value.validate()) {
        return Promise.reject()
      }
      if (!activeColumn.value) {
        return Promise.reject()
      }

      props.monitor.params.column = activeColumn.value.name
      props.monitor.params.columnUnit = activeColumn.value.unit

      if (props.monitor.id) {
        return monitorMan.updateMetricMonitor(props.monitor)
      }
      return monitorMan.createMetricMonitor(props.monitor)
    }

    function formatNum(n: number) {
      return numbro(n).format({
        mantissa: mantissa(n),
        trimMantissa: true,
      })
    }

    function mantissa(n: number) {
      if (n < 0.1) {
        return 3
      }
      if (n < 1) {
        return 2
      }
      if (n < 10) {
        return 1
      }
      return 0
    }

    return {
      eventBus,
      channels,

      form,
      isValid,
      rules,
      forMinuteItems,
      submit,

      uql,
      monitorMan,
      activeMetrics,
      axiosParams,
      timeseries,
      styledTimeseries,
      activeColumn,
      observedMin,
      observedMax,
      observedAvg,
      formatNum,
    }
  },
})
</script>

<style lang="scss" scoped></style>
