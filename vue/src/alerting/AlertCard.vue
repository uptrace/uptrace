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
                  {{ alert.data.status === AlertStatus.Open ? 'Close' : 'Reopen' }} alert
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
                  {{ alert.data.status === AlertStatus.Open ? 'Close' : 'Reopen' }} alert
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
import { useProject } from '@/org/use-projects'
import { useAlert, useAlertManager, AlertType, AlertStatus } from '@/alerting/use-alerts'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import AlertCardSpan from '@/alerting/AlertCardSpan.vue'
import AlertCardMetric from '@/alerting/AlertCardMetric.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'AlertCard',
  components: { FixedDateRangePicker, AlertCardSpan, AlertCardMetric },

  props: {
    alertId: {
      type: Number,
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
    const project = useProject()

    const alert = useAlert(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/projects/${projectId}/alerts/${props.alertId}`,
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
        text: project.data?.name ?? 'Project',
        to: {
          name: 'ProjectShow',
        },
      })

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

      bs.push({
        text: truncate(alert.data.name, { length: 60 }),
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
      AlertStatus,
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
