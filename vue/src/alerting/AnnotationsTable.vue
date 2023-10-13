<template>
  <v-data-table
    :headers="headers"
    :loading="loading"
    :items="annotations"
    hide-default-footer
    no-data-text="There are no annotations"
    class="v-data-table--narrow"
    :sort-by.sync="order.column"
    :sort-desc.sync="order.desc"
    must-sort
    @update:sort-by="$nextTick(() => (order.desc = true))"
  >
    <template #item="{ item }">
      <AnnotationsTableRow :annotation="item" @change="$emit('change', $event)" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { Annotation } from '@/org/use-annotations'
import { UseOrder } from '@/use/order'

// Components
import AnnotationsTableRow from '@/alerting/AnnotationsTableRow.vue'

export default defineComponent({
  name: 'AnnotationsTable',
  components: { AnnotationsTableRow },

  props: {
    annotations: {
      type: Array as PropType<Annotation[]>,
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
  },

  setup(props) {
    const headers = computed(() => {
      const headers = []
      headers.push({ text: 'Name', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'Attributes', value: 'attrs', sortable: false, align: 'start' })
      headers.push({ text: 'Actions', value: 'actions', sortable: false, align: 'center' })
      headers.push({ text: 'Created at', value: 'createdAt', sortable: true, align: 'end' })
      return headers
    })

    return {
      headers,
    }
  },
})
</script>

<style lang="scss" scoped></style>
