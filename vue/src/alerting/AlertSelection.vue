<template>
  <div class="d-flex">
    <v-simple-checkbox
      :value="selection.isFullPageSelected"
      :indeterminate="selection.alertsOnPage.length > 0 && !selection.isFullPageSelected"
      :ripple="false"
      @click="selection.togglePage"
    ></v-simple-checkbox>

    <v-btn depressed small class="ml-3" @click="selection.toggleAll">
      <span> {{ selection.isAllSelected ? 'Deselect all' : 'Select all' }}</span>
      <span v-if="selection.alerts.length">({{ selection.alerts.length }})</span>
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
    <v-btn
      v-if="selection.alerts.length"
      :loading="alertMan.pending"
      depressed
      small
      class="ml-3"
      @click="deleteAlerts"
    >
      Delete ({{ selection.alerts.length }})
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

    function openAlerts() {
      alertMan.open(props.selection.closedAlerts).then(() => {
        props.selection.reset()
        ctx.emit('change')
      })
    }

    function closeAlerts() {
      alertMan.close(props.selection.openAlerts).then(() => {
        props.selection.reset()
        ctx.emit('change')
      })
    }

    function deleteAlerts() {
      confirm
        .open('Delete', 'Do you really want to delete selected alerts?')
        .then(() => {
          alertMan.delete(props.selection.alerts).then(() => {
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
      deleteAlerts,
    }
  },
})
</script>

<style lang="scss" scoped></style>
