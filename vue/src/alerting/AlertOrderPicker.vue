<template>
  <v-select v-model="order" dense solo flat filled hide-details mandatory :items="items">
    <template #prepend-inner>
      <v-icon>mdi-sort-descending</v-icon>
    </template>
  </v-select>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch } from 'vue'

export default defineComponent({
  name: 'AlertsOrderPicker',

  props: {
    value: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const order = shallowRef(props.value)
    const items = [
      {
        text: 'Updated at',
        value: 'updated_at',
      },
      {
        text: 'Created at',
        value: 'created_at',
      },
    ]

    watch(order, (order) => ctx.emit('input', order), { immediate: true })

    return {
      order,
      items,
    }
  },
})
</script>

<style lang="scss" scoped></style>
