<template>
  <v-card flat>
    <div v-if="!items.length" class="py-4 text-body-2 text--secondary text-center">
      No matching values
    </div>

    <v-list v-else dense class="py-0">
      <v-list-item v-for="item in pagedItems" :key="item.value" @click="resetItem(item)">
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="values.indexOf(item.value) >= 0"
            dense
            @click.stop
            @change="selectItem(item, $event)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content class="text-truncate">
          <v-list-item-title>{{ item.value }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action v-if="item.count" class="my-0">
          <v-list-item-action-text><XNum :value="item.count" /></v-list-item-action-text>
        </v-list-item-action>
      </v-list-item>
    </v-list>

    <XPagination v-if="pager.numPage > 1" :pager="pager" total-visible="5" :show-pager="false" />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref, computed, watch, PropType } from 'vue'

// Composables
import { usePager } from '@/use/pager'
import { Item } from '@/components/facet/types'

export default defineComponent({
  name: 'SpanFacetBody',

  props: {
    value: {
      type: Array as PropType<string[]>,
      default: undefined,
    },
    items: {
      type: Array as PropType<Item[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const pager = usePager({ perPage: 10 })
    const values = ref<string[]>([])

    const pagedItems = computed(() => {
      return props.items.slice(pager.pos.start, pager.pos.end)
    })

    watch(
      () => props.value,
      (value) => {
        values.value = value ?? []
      },
      { immediate: true },
    )

    watch(
      () => props.items.length,
      (numItem) => {
        pager.numItem = numItem
      },
      { immediate: true },
    )

    function selectItem(item: Item, selected: boolean) {
      if (selected) {
        values.value.push(item.value)
      } else {
        const index = values.value.indexOf(item.value)
        if (index >= 0) {
          values.value.splice(index, 1)
        }
      }
      ctx.emit('input', values.value)
    }

    function resetItem(item: Item) {
      const index = values.value.indexOf(item.value)
      if (index >= 0) {
        values.value.splice(index, 1)
      } else {
        values.value = [item.value]
      }
      ctx.emit('input', values.value)
      ctx.emit('click:close')
    }

    return { pager, values, pagedItems, selectItem, resetItem }
  },
})
</script>

<style lang="scss" scoped></style>
