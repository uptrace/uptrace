<template>
  <v-container fluild>
    <v-row>
      <v-col>
        <SinglePanel title="Monitor options" expanded>
          <v-text-field
            v-model="monitor.name"
            label="Monitor name"
            filled
            dense
            :rules="rules.name"
          />

          <PanelSection title="Time offset">
            <v-text-field
              v-model.number="timeOffset"
              type="number"
              hint="Use a negative number to adjuct check time"
              placeholder="-60"
              suffix="minutes"
              persistent-hint
              filled
              dense
              :rules="rules.timeOffset"
            />
          </PanelSection>
        </SinglePanel>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <SinglePanel title="Trigger conditions" expanded>
          <PanelSection title="Grouping interval">
            <v-select
              v-model="monitor.params.groupingInterval"
              :items="groupingIntervalItems"
              hint="Use larger grouping interval for infrequently updated metrics"
              persistent-hint
              filled
              dense
              hide-details="auto"
            />
          </PanelSection>

          <PanelSection title="Check the last">
            <v-select
              v-model="monitor.params.checkNumPoint"
              hint="Create an alert if the last N points are outside of the allowed range"
              persistent-hint
              :items="checkNumPointItems"
              filled
              dense
            />
          </PanelSection>

          <PanelSection>
            <template #title>
              <span>Min allowed value</span>
              <span v-if="observedMin" class="ml-1">
                (observed min:
                <strong>{{ numVerbose(observedMin) }}</strong
                >)
              </span>
            </template>
            <v-text-field
              v-model.number="monitor.params.minValue"
              type="number"
              :suffix="activeColumn?.unit"
              hint="Leave empty to disable"
              persistent-hint
              filled
              dense
              clearable
              :rules="rules.minValue"
            />
          </PanelSection>

          <PanelSection>
            <template #title>
              <span>Max allowed value</span>
              <span v-if="observedMax" class="ml-1">
                (observed max:
                <strong>{{ numVerbose(observedMax) }}</strong
                >)
              </span>
            </template>
            <v-text-field
              v-model.number="monitor.params.maxValue"
              type="number"
              :suffix="activeColumn?.unit"
              hint="Leave empty to disable"
              persistent-hint
              filled
              dense
              clearable
              :rules="rules.maxValue"
            />
          </PanelSection>
        </SinglePanel>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <SinglePanel title="Notifications" expanded>
          <v-checkbox
            v-model="monitor.notifyEveryoneByEmail"
            label="Notify everyone by email"
            hide-details="auto"
          />

          <PanelSection title="Notification channels, e.g. Slack, Telegram, etc.">
            <v-select
              v-model="monitor.channelIds"
              multiple
              menu-props="offsetY"
              persistent-hint
              filled
              dense
              :items="channels.items"
              item-text="name"
              item-value="id"
              hide-details="auto"
            >
              <template #item="{ item }">
                <v-list-item-action class="my-0 mr-4">
                  <v-simple-checkbox :value="monitor.channelIds.includes(item.id)">
                  </v-simple-checkbox>
                </v-list-item-action>
                <v-list-item-content>
                  <v-list-item-title>{{ item.name }} ({{ item.type }})</v-list-item-title>
                </v-list-item-content>
              </template>
            </v-select>
          </PanelSection>
        </SinglePanel>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <SinglePanel title="Flapping monitors">
          <p class="text--secondary">
            Flapping alerts occur when the monitor triggers the same alert in a short period of
            time. To reduce the noise, you can re-evaluate the alert thresholds or configure
            additional conditions required to close the alert.
          </p>

          <p class="text--secondary">
            For example, the filesystem utilization monitor can go back and forth from
            <code>0.89</code> to <code>0.9</code>, causing the alert status to constantly change. By
            configuring the minimum deviation value to <code>0.1</code>, the alert won't be closed
            until the utilization value changes from <code>0.9</code> to <code>0.8</code>.
          </p>

          <v-checkbox
            v-model="monitor.params.flapping"
            label="Enable flapping alerts mitigations"
            class="mt-0"
          />

          <PanelSection>
            <template #title>
              <span>Min change in value to trigger an alert</span>
              <span v-if="observedAvg" class="ml-1">
                (observed average:
                <strong>{{ numVerbose(observedAvg) }}</strong
                >)
              </span>
            </template>

            <v-text-field
              v-model.number="monitor.params.minDevValue"
              type="number"
              placeholder="Number"
              :suffix="activeColumn?.unit"
              filled
              dense
              :rules="rules.minDevValue"
              hide-details="auto"
              :disabled="!monitor.params.flapping"
            />
          </PanelSection>

          <PanelSection title="Min deviation fraction">
            <v-text-field
              v-model.number="monitor.params.minDevFraction"
              type="number"
              placeholder="Number"
              hint="0.2 means that value must change by at least 20%"
              persistent-hint
              filled
              dense
              :rules="rules.minDevFraction"
              :disabled="!monitor.params.flapping"
            />
          </PanelSection>
        </SinglePanel>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { formatDuration } from 'date-fns'
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { useNotifChannels } from '@/alerting/use-notif-channels'

// Components
import SinglePanel from '@/components/SinglePanel.vue'
import PanelSection from '@/components/PanelSection.vue'

// Misc
import { Timeseries, MetricColumn } from '@/metrics/types'
import { MetricMonitor } from '@/alerting/types'
import { requiredRule, minMaxRule } from '@/util/validation'
import { inflect, numVerbose } from '@/util/fmt'
import { MINUTE, HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'MonitorMetricFormOptions',
  components: {
    SinglePanel,
    PanelSection,
  },

  props: {
    monitor: {
      type: Object as PropType<MetricMonitor>,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      required: true,
    },
    timeseries: {
      type: Array as PropType<Timeseries[]>,
      required: true,
    },
    form: {
      type: Object,
      default: undefined,
    },
  },

  setup(props, ctx) {
    const channels = useNotifChannels(() => {
      return {}
    })
    const timeOffset = computed({
      get() {
        return props.monitor.params.timeOffset / MINUTE
      },
      set(minutes: number) {
        props.monitor.params.timeOffset = minutes * MINUTE
      },
    })

    const rules = {
      name: [requiredRule],
      boundsSource: [requiredRule],
      minValue: [
        (v: any) => {
          if (
            typeof props.monitor.params.minValue !== 'number' &&
            typeof props.monitor.params.maxValue !== 'number'
          ) {
            return 'At least min or max value is required'
          }
          return true
        },
      ],
      maxValue: [
        (v: any) => {
          if (
            typeof props.monitor.params.minValue !== 'number' &&
            typeof props.monitor.params.maxValue !== 'number'
          ) {
            return 'At least min or max value is required'
          }
          if (
            typeof props.monitor.params.minValue !== 'number' ||
            typeof props.monitor.params.maxValue !== 'number'
          ) {
            return true
          }
          if (props.monitor.params.maxValue < props.monitor.params.minValue) {
            return 'Max value should be greater than or equal min'
          }
          return true
        },
      ],
      timeOffset: [minMaxRule(-300, 300)],
    }
    const checkNumPointItems = computed(() => {
      const maxDuration = 24 * HOUR

      const items = []

      for (let n of [1, 3, 5, 10, 15]) {
        const duration = n * MINUTE
        if (duration > maxDuration) {
          break
        }

        const noun = inflect(n, 'point', 'points')
        const hours = Math.trunc(duration / HOUR)
        const minutes = Math.trunc((duration - hours * HOUR) / MINUTE)
        const durationStr = formatDuration({ hours, minutes })

        items.push({
          text: `${n} ${noun} (${durationStr})`,
          value: n,
        })
      }

      return items
    })

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
    watch(
      activeColumn,
      (activeColumn) => {
        props.monitor.params.column = activeColumn?.name ?? ''
        props.monitor.params.columnUnit = activeColumn?.unit ?? ''
      },
      { immediate: true },
    )

    const observedMin = computed(() => {
      let min = Number.MAX_VALUE

      for (let ts of props.timeseries) {
        if (ts.min === null) {
          continue
        }
        if (ts.min < min) {
          min = ts.min
        }
      }

      if (min !== Number.MAX_VALUE) {
        return min
      }
      return undefined
    })

    const observedMax = computed(() => {
      let max = Number.MIN_VALUE

      for (let ts of props.timeseries) {
        if (ts.max === null) {
          continue
        }
        if (ts.max > max) {
          max = ts.max
        }
      }

      if (max !== Number.MIN_VALUE) {
        return max
      }
      return undefined
    })

    const observedAvg = computed(() => {
      let sum = 0
      let count = 0

      for (let ts of props.timeseries) {
        for (let num of ts.value) {
          if (num !== null) {
            sum += num
            count++
          }
        }
      }

      if (count) {
        return sum / count
      }
      return 0
    })

    watch(
      () => props.monitor.params.minValue,
      () => props.form?.validate(),
    )
    watch(
      () => props.monitor.params.maxValue,
      () => props.form?.validate(),
    )

    return {
      channels,
      timeOffset,

      rules,
      checkNumPointItems,

      observedMin,
      observedMax,
      observedAvg,
      activeColumn,
      numVerbose,
    }
  },
})
</script>

<style lang="scss" scoped></style>
