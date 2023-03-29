<template>
  <v-data-table
    :headers="headers"
    :loading="loading"
    :items="channels"
    :hide-default-footer="channels.length <= 10"
    no-data-text="There are no notification channels"
    class="v-data-table--narrow"
  >
    <template #item="{ item }">
      <NotifChannelsTableRow
        :channel="item"
        @change="$emit('change', $event)"
      ></NotifChannelsTableRow>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { NotifChannel } from '@/alerting/use-notif-channels'

import NotifChannelsTableRow from '@/alerting/NotifChannelsTableRow.vue'

export default defineComponent({
  name: 'NotifChannelsTable',
  components: { NotifChannelsTableRow },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    channels: {
      type: Array as PropType<NotifChannel[]>,
      required: true,
    },
  },

  setup() {
    const headers = computed(() => {
      const headers = []
      headers.push({ text: 'Name', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'Type', value: 'type', sortable: true, align: 'start' })
      headers.push({ text: 'State', value: 'state', sortable: true, align: 'center' })
      headers.push({ text: 'Actions', value: 'actions', sortable: false, align: 'center' })
      return headers
    })

    return { headers }
  },
})
</script>

<style lang="scss" scoped></style>
