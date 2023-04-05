<template>
  <div class="d-flex">
    <v-simple-checkbox
      :value="selection.alerts.length > 0"
      :ripple="false"
      @click="selection.toggleAll"
    ></v-simple-checkbox>

    <v-btn
      v-if="selection.hasOpen && !selection.alerts.length"
      :loading="alertMan.pending"
      depressed
      small
      class="ml-3"
      @click="closeAllAlerts"
    >
      Close all
    </v-btn>
    <v-btn
      v-if="selection.openAlerts.length"
      :loading="alertMan.pending"
      depressed
      small
      class="ml-3"
      @click="closeAlerts"
    >
      Close ({{ selection.openAlerts.length }})
    </v-btn>
    <v-btn
      v-if="selection.closedAlerts.length"
      :loading="alertMan.pending"
      depressed
      small
      class="ml-3"
      @click="openAlerts"
    >
      Open ({{ selection.closedAlerts.length }})
    </v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import { useAlertManager, UseAlertSelection } from '@/alerting/use-alerts'

export default defineComponent({
  name: 'AlertsSelection',

  props: {
    selection: {
      type: Object as PropType<UseAlertSelection>,
      required: true,
    },
  },

  setup(props, ctx) {
    const confirm = useConfirm()
    const alertMan = useAlertManager()

    const closeAlerts = function () {
      alertMan.close(props.selection.openAlerts).then(() => {
        props.selection.reset()
        ctx.emit('change')
      })
    }

    const openAlerts = function () {
      alertMan.open(props.selection.closedAlerts).then(() => {
        props.selection.reset()
        ctx.emit('change')
      })
    }

    const closeAllAlerts = function () {
      confirm
        .open('Close all', 'Do you really want to close all alerts?')
        .then(() => {
          alertMan.closeAll().then(() => {
            props.selection.reset()
            ctx.emit('change')
          })
        })
        .catch(() => {})
    }

    return {
      alertMan,
      closeAlerts,
      openAlerts,
      closeAllAlerts,
    }
  },
})
</script>

<style lang="scss" scoped></style>
