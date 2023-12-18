<template>
  <v-card flat>
    <div v-if="!items.length" class="py-4 text-body-2 text--secondary text-center">
      No matching values
    </div>

    <v-list v-else dense class="py-0">
      <v-list-item v-for="item in pagedItems" :key="item.value" @click="toggleOne(item.value)">
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="values.includes(item.value)"
            @click.stop="toggle(item.value)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content class="text-truncate">
          <v-list-item-title>{{ item.value }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action v-if="item.count" class="my-0">
          <v-list-item-action-text><NumValue :value="item.count" /></v-list-item-action-text>
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
  name: 'FacetItemBody',

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
    const pager = usePager()
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

    function toggle(itemValue: string) {
      const selected = values.value.slice()
      const index = selected.indexOf(itemValue)
      if (index >= 0) {
        selected.splice(index, 1)
      } else {
        selected.push(itemValue)
      }
      ctx.emit('input', selected)
    }

    function toggleOne(itemValue: string) {
      let selected = [itemValue]
      if (values.value.length === 1 && values.value.includes(itemValue)) {
        selected = []
      }
      ctx.emit('input', selected)
    }

    return { pager, values, pagedItems, toggle, toggleOne }
  },
})
</script>

<style lang="scss" scoped></style>
