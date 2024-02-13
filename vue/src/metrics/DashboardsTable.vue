<template>
  <v-data-table
    :headers="headers"
    :loading="loading"
    :items="dashboards"
    :items-per-page="itemsPerPage"
    :hide-default-footer="dashboards.length <= itemsPerPage"
    no-data-text="There are no dashboards"
    class="v-data-table--narrow"
    :custom-sort="(items) => items"
    :sort-by.sync="order.column"
    :sort-desc.sync="order.desc"
    @update:sort-by="$nextTick(() => (order.desc = true))"
  >
    <template #item="{ item }">
      <DashboardsTableRow :dashboard="item" :headers="headers" @change="$emit('change', $event)" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'

// Components
import DashboardsTableRow from '@/metrics/DashboardsTableRow.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardsTable',
  components: { DashboardsTableRow },

  props: {
    dashboards: {
      type: Array as PropType<Dashboard[]>,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    itemsPerPage: {
      type: Number,
      default: 100,
    },
  },

  setup(props) {
    const headers = computed(() => {
      const headers = []
      headers.push({ text: 'Dashboard Name', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'Updated at', value: 'updatedAt', sortable: true, align: 'start' })
      headers.push({ text: 'Actions', value: 'actions', sortable: false, align: 'end' })
      return headers
    })

    return { headers }
  },
})
</script>

<style lang="scss" scoped></style>
