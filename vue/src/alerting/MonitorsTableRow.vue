<template>
  <tr class="cursor-pointer" @click="$router.push(monitorRouteFor(monitor))">
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
    <td class="text-center text-no-wrap">
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

      <v-btn :to="monitorRouteFor(monitor)" icon title="Edit monitor" @click.stop
        ><v-icon>mdi-pencil-outline</v-icon></v-btn
      >
      <v-btn :loading="monitorMan.pending" icon title="Delete monitor" @click.stop="deleteMonitor"
        ><v-icon>mdi-delete-outline</v-icon></v-btn
      >
    </td>
    <td class="text-center">
      <router-link
        v-if="monitor.alertCount"
        :to="{
          name: 'AlertList',
          query: {
            q: 'monitor:' + monitor.id,
            state: null,
          },
        }"
        class="link"
        >{{ monitor.alertCount }} alerts</router-link
      >
    </td>
    <td>
      <XDate v-if="monitor.updatedAt" :date="monitor.updatedAt" format="relative" />
    </td>
  </tr>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import {
  useMonitorManager,
  monitorRouteFor,
  Monitor,
  MonitorType,
  MonitorState,
} from '@/alerting/use-monitors'

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
      monitorRouteFor,

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
