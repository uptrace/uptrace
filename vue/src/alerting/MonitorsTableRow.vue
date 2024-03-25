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
      <v-chip
        v-if="monitor.alertOpenCount"
        :to="routeForOpenAlerts"
        label
        color="red lighten-5"
        light
      >
        <span class="font-weight-medium">{{ monitor.alertOpenCount }}</span>
        <span class="ml-1"> open</span>
      </v-chip>
      <v-chip
        v-if="monitor.alertClosedCount"
        :to="routeForClosedAlerts"
        label
        color="green lighten-5"
        light
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

      <v-btn :to="routeForMonitor(monitor)" icon title="Edit monitor" @click.stop>
        <v-icon>mdi-pencil-outline</v-icon>
      </v-btn>
      <v-btn :loading="monitorMan.pending" icon title="Delete monitor" @click.stop="deleteMonitor">
        <v-icon>mdi-delete-outline</v-icon>
      </v-btn>
      <v-menu v-model="menu" offset-y>
        <template #activator="{ on: onMenu, attrs }">
          <v-btn icon v-bind="attrs" v-on="onMenu">
            <v-icon>mdi-dots-vertical</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="openMonitorDialog(monitor)">
            <v-list-item-icon>
              <v-icon>mdi-eye</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>View Yaml</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
    </td>

    <MonitorYamlDialog v-if="dialog" v-model="dialog" :monitor="activeMonitor" />
  </tr>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import { useMonitorManager, routeForMonitor } from '@/alerting/use-monitors'
import { AlertStatus } from '@/alerting/use-alerts'

// Components
import MonitorTypeIcon from '@/alerting/MonitorTypeIcon.vue'
import MonitorStateAvatar from '@/alerting/MonitorStateAvatar.vue'
import MonitorYamlDialog from '@/alerting/MonitorYamlDialog.vue'

// Misc
import { Monitor, MonitorType, MonitorState } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorsTableRow',
  components: { MonitorTypeIcon, MonitorStateAvatar, MonitorYamlDialog },

  props: {
    monitor: {
      type: Object as PropType<Monitor>,
      required: true,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)
    const activeMonitor = shallowRef<Monitor>()
    const dialog = shallowRef(false)
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
          'attrs.alert_status': status,
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

    function openMonitorDialog(monitor: Monitor) {
      activeMonitor.value = monitor
      dialog.value = true
    }

    return {
      MonitorType,
      MonitorState,

      menu,
      dialog,
      activeMonitor,
      routeForMonitor,
      routeForOpenAlerts,
      routeForClosedAlerts,

      monitorMan,
      activateMonitor,
      pauseMonitor,
      deleteMonitor,
      openMonitorDialog,
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
