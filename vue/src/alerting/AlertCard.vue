<template>
  <div>
    <v-skeleton-loader v-if="!alert.data" type="article, table" />

    <template v-else>
      <PageToolbar :fluid="fluid" :loading="alert.loading">
        <v-breadcrumbs :items="breadcrumbs" divider=">" large></v-breadcrumbs>

        <v-spacer />

        <FixedDateRangePicker :date-range="dateRange" :around="alert.data.updatedAt" show-reload />
      </PageToolbar>

      <v-container :fluid="fluid" class="py-4">
        <v-row>
          <v-col>
            <AlertCardSpan
              v-if="alert.data.type === AlertType.Error"
              :date-range="dateRange"
              :alert="alert.data"
            >
              <template slot="append-action">
                <v-btn
                  :loading="alertMan.pending"
                  depressed
                  small
                  class="ml-2"
                  @click="toggleAlert"
                >
                  {{ alert.data.state === AlertState.Open ? 'Close' : 'Reopen' }} alert
                </v-btn>
              </template>
            </AlertCardSpan>

            <AlertCardMetric v-else :date-range="dateRange" :alert="alert.data">
              <template slot="append-action">
                <v-btn
                  :loading="alertMan.pending"
                  depressed
                  small
                  class="ml-2"
                  @click="toggleAlert"
                >
                  {{ alert.data.state === AlertState.Open ? 'Close' : 'Reopen' }} alert
                </v-btn>
              </template>
            </AlertCardMetric>
          </v-col>
        </v-row>
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { truncate } from 'lodash-es'
import { defineComponent, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { useDateRange } from '@/use/date-range'
import { useAlert, useAlertManager, AlertType, AlertState } from '@/alerting/use-alerts'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import AlertCardSpan from '@/alerting/AlertCardSpan.vue'
import AlertCardMetric from '@/alerting/AlertCardMetric.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'

export default defineComponent({
  name: 'AlertCard',
  components: { FixedDateRangePicker, AlertCardSpan, AlertCardMetric },

  props: {
    alertId: {
      type: String,
      required: true,
    },
    fluid: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const dateRange = useDateRange()

    const alert = useAlert(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/projects/${projectId}/alerts/${props.alertId}`,
      }
    })
    const alertMan = useAlertManager()

    useTitle(
      computed(() => {
        return alert.data?.name ?? 'Alert'
      }),
    )

    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: 'Alerts',
        to: {
          name: 'AlertList',
        },
        exact: true,
      })

      if (!alert.data) {
        bs.push({ text: 'Alert' })
        return bs
      }

      const system = alert.data.attrs[AttrKey.spanSystem]
      if (system) {
        bs.push({
          text: system,
          to: {
            name: isEventSystem(system) ? 'EventGroupList' : 'SpanGroupList',
            query: {
              ...dateRange.queryParams(),
              system,
            },
          },
          exact: true,
        })
      }

      bs.push({
        text: truncate(alert.data.name, { length: 80 }),
        to: {
          name: 'AlertShow',
          params: { alertId: alert.data.id },
        },
      })

      return bs
    })

    function toggleAlert() {
      if (!alert.data) {
        return
      }
      alertMan.toggle(alert.data).then(() => {
        alert.reload()
        ctx.emit('change', alert)
      })
    }

    return {
      AlertType,
      AlertState,
      AttrKey,

      dateRange,
      breadcrumbs,

      alert,
      alertMan,

      toggleAlert,
    }
  },
})
</script>

<style lang="scss" scoped></style>
