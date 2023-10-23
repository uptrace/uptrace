<template>
  <tr class="cursor-pointer" @click="$router.push(routeForMonitor(monitor))">
    <td>
      {{ monitor.name }}
    </td>
    <td class="text-no-wrap">
      <MonitorTypeIcon :type="monitor.type" class="mr-2" />
      <span>{{ monitor.type }}</span>
    </td>
    <td class="text-center">
      <MonitorStateAvatar :state="monitor.state" />
    </td>
    <td class="text-center">
      <v-chip v-if="monitor.alertOpenCount" :to="routeForOpenAlerts" label color="red lighten-5">
        <span class="font-weight-medium">{{ monitor.alertOpenCount }}</span>
        <span class="ml-1"> open</span>
      </v-chip>
      <v-chip
        v-if="monitor.alertClosedCount"
        :to="routeForClosedAlerts"
        label
        color="green lighten-5"
        class="ml-2"
      >
        <span class="font-weight-medium">{{ monitor.alertClosedCount }}</span>
        <span class="ml-1"> closed</span>
      </v-chip>
    </td>
    <td>
      <DateValue v-if="monitor.updatedAt" :value="monitor.updatedAt" format="relative" />
    </td>
    <td class="text-right text-no-wrap">
      <v-btn
        v-if="monitor.state != MonitorState.Paused"
        icon
        title="Pause monitor"
        @click.stop="pauseMonitor(monitor)"
      >
        <v-icon>mdi-pause</v-icon>
      </v-btn>
      <v-btn v-else icon title="Resume monitor" @click.stop="activateMonitor(monitor)">
        <v-icon>mdi-play</v-icon>
      </v-btn>

      <v-btn :to="routeForMonitor(monitor)" icon title="Edit monitor" @click.stop
        ><v-icon>mdi-pencil-outline</v-icon></v-btn
      >
      <v-btn :loading="monitorMan.pending" icon title="Delete monitor" @click.stop="deleteMonitor"
        ><v-icon>mdi-delete-outline</v-icon></v-btn
      >
    </td>
  </tr>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import {
  useMonitorManager,
  routeForMonitor,
  Monitor,
  MonitorType,
  MonitorState,
} from '@/alerting/use-monitors'
import { AlertStatus } from '@/alerting/use-alerts'

// Components
import MonitorTypeIcon from '@/alerting/MonitorTypeIcon.vue'
import MonitorStateAvatar from '@/alerting/MonitorStateAvatar.vue'

export default defineComponent({
  name: 'MonitorsTableRow',
  components: { MonitorTypeIcon, MonitorStateAvatar },

  props: {
    monitor: {
      type: Object as PropType<Monitor>,
      required: true,
    },
  },

  setup(props, ctx) {
    const confirm = useConfirm()
    const monitorMan = useMonitorManager()

    const routeForOpenAlerts = computed(() => {
      return routeForAlerts(AlertStatus.Open)
    })

    const routeForClosedAlerts = computed(() => {
      return routeForAlerts(AlertStatus.Closed)
    })

    function routeForAlerts(status: AlertStatus) {
      return {
        name: 'AlertList',
        query: {
          q: 'monitor:' + props.monitor.id,
          'attrs.alert.status': status,
        },
      }
    }

    function activateMonitor() {
      monitorMan.activate(props.monitor).then(() => {
        ctx.emit('change', props.monitor)
      })
    }

    function pauseMonitor(monitor: Monitor) {
      monitorMan.pause(monitor).then(() => {
        ctx.emit('change', props.monitor)
      })
    }

    function deleteMonitor() {
      confirm
        .open('Delete monitor', `Do you really want to delete "${props.monitor.name}" monitor?`)
        .then(() => monitorMan.del(props.monitor))
        .then((monitor) => ctx.emit('change', monitor))
        .catch(() => {})
    }

    return {
      MonitorType,
      MonitorState,

      routeForMonitor,
      routeForOpenAlerts,
      routeForClosedAlerts,

      monitorMan,
      activateMonitor,
      pauseMonitor,
      deleteMonitor,
    }
  },
})
</script>

<style lang="scss" scoped>
.link {
  text-decoration: none;
  font-weight: 500;

  &:hover {
    text-decoration: underline;
  }
}
</style>
