<template>
  <v-data-table
    :headers="headers"
    :loading="loading"
    :items="monitors"
    :hide-default-footer="monitors.length <= 10"
    no-data-text="There are no monitors"
    class="v-data-table--narrow"
  >
    <template #item="{ item }">
      <MonitorsTableRow :monitor="item" @change="$emit('change', $event)" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Components
import MonitorsTableRow from '@/alerting/MonitorsTableRow.vue'

// Misc
import { Monitor } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorsTable',
  components: { MonitorsTableRow },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    monitors: {
      type: Array as PropType<Monitor[]>,
      required: true,
    },
  },

  setup() {
    const headers = computed(() => {
      const headers = []
      headers.push({ text: 'Monitor Name', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'Type', value: 'type', sortable: true, align: 'start' })
      headers.push({ text: 'State', value: 'state', sortable: true, align: 'center' })
      headers.push({ text: 'Alerts', value: 'alertCount', sortable: true, align: 'center' })
      headers.push({ text: 'Last activity at', value: 'updatedAt', sortable: true, align: 'start' })
      headers.push({ text: 'Actions', value: 'actions', sortable: false, align: 'end' })
      return headers
    })

    return { headers }
  },
})
</script>

<style lang="scss" scoped></style>
