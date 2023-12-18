<template>
  <div ref="gridStackEl" class="grid-stack">
    <div
      v-for="item in rowItems"
      :id="`gs-item-${item.id}`"
      :key="item.id"
      :gs-id="item.id"
      :gs-w="item.width"
      :gs-h="item.height"
      :gs-x="item.xAxis"
      :gs-y="item.yAxis"
      :gs-auto-position="gsAutoPosition"
      class="grid-stack-item"
    >
      <div class="grid-stack-item-content">
        <slot
          name="item"
          v-bind="{
            attrs: { gridItem: item, height: itemHeight(item) },
            on: {
              ready() {
                resizeItem(item)
              },
            },
          }"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import 'gridstack/dist/gridstack.min.css'
import 'gridstack/dist/gridstack-extra.min.css'
import {
  GridStack,
  GridStackOptions,
  GridStackWidget,
  GridStackNode,
  GridItemHTMLElement,
  Utils,
} from 'gridstack'
import axios from 'axios'

import {
  defineComponent,
  shallowRef,
  ref,
  set,
  del,
  computed,
  watch,
  onMounted,
  onBeforeUnmount,
  PropType,
} from 'vue'

// Composables
import { useRoute } from '@/use/router'

// Misc
import { GridItem, GridItemType, GridItemPos } from '@/metrics/types'

export default defineComponent({
  name: 'GridStackCard',

  props: {
    items: {
      type: Array as PropType<GridItem[]>,
      required: true,
    },
    rowId: {
      type: Number,
      default: null,
    },
    rowIndex: {
      type: Number,
      default: 0,
    },
  },

  setup(props, ctx) {
    const route = useRoute()

    const gridStackEl = shallowRef()
    let gridStack: GridStack | undefined

    const rowItems = computed(() => {
      const rowId = props.rowId ?? 0
      return props.items.filter((item) => {
        return item.rowId === rowId
      })
    })

    const itemsHeight = ref<Record<number, number>>({})
    const cellHeight = 10
    function itemHeight(gridItem: GridItem): number {
      const height = itemsHeight.value[gridItem.id]
      if (height) {
        return height * cellHeight
      }
      return gridItem.height * cellHeight
    }

    //------------------------------------------------------------------------------

    function saveItemsLayout(items: GridStackNode[]) {
      const layout: GridItemPos[] = []

      for (let node of items) {
        const id = parseInt(node.id!, 10)
        if (!id) {
          continue
        }

        layout.push({
          id,
          width: node.w || 0,
          height: node.h || 0,
          xAxis: node.x || 0,
          yAxis: node.y || 0,
        })
      }

      const { projectId, dashId } = route.value.params
      const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid/layout`
      return axios
        .request({
          method: 'PUT',
          url,
          data: {
            items: layout,
            rowId: props.rowId,
          },
        })
        .then(() => {
          return layout
        })
    }

    //------------------------------------------------------------------------------

    const gsAutoPosition = computed(() => {
      const auto = rowItems.value.every((item) => item.xAxis === 0 && item.yAxis === 0)
      return auto
    })

    const gridStackOptions = computed((): GridStackOptions => {
      const options = {
        animate: false,
        column: 12,
        cellHeight,
        columnOpts: { breakpoints: [{ w: 960, c: 1 }] },
        margin: 3,
        minRow: 8,
        draggable: {
          handle: '.drag-handle',
        },
        resizable: { handles: 'se,sw' },
        acceptWidgets: true,
      }
      return options
    })

    onMounted(() => {
      watch(gridStackOptions, updateGridStack, { immediate: true })
      watch(() => rowItems.value, updateGridStack)
      watch(() => props.rowIndex, updateGridStack)
    })

    onBeforeUnmount(() => {
      if (gridStack) {
        gridStack.destroy(false)
        gridStack = undefined
      }
    })

    function updateGridStack() {
      if (gridStack) {
        gridStack.setAnimation(false)
        gridStack.destroy(false)
        gridStack = undefined
      }

      gridStack = GridStack.init(gridStackOptions.value, gridStackEl.value)
      setTimeout(() => {
        gridStack?.setAnimation(true)
      }, 1000)

      const items = gridStack.getGridItems()
      for (let el of items) {
        const node = el.gridstackNode
        if (!node) {
          continue
        }

        const id = parseInt(node.id!, 10)
        const found = rowItems.value.find((item) => item.id === id)
        if (found) {
          const opts: GridStackWidget = {
            w: found.width,
            h: found.height,
          }
          if (!gsAutoPosition.value) {
            opts.x = found.xAxis
            opts.y = found.yAxis
          }
          switch (found.type) {
            case GridItemType.Gauge:
            case GridItemType.Heatmap:
            case GridItemType.Table:
              opts.sizeToContent = true
          }
          gridStack.update(el, opts)
        }
      }

      gridStack.on('added', (_event: unknown, items: GridStackNode[]) => {
        for (let node of items) {
          const id = parseInt(node.id!, 10)
          set(itemsHeight.value, id, node.h)
        }
        saveItemsLayout(items)
      })
      gridStack.on('removed', (_event: unknown, items: GridStackNode[]) => {
        for (let node of items) {
          const id = parseInt(node.id!, 10)
          del(itemsHeight.value, id)
        }
      })
      gridStack.on('change', (_event: unknown, items: GridStackNode[]) => {
        for (let node of items) {
          const id = parseInt(node.id!, 10)
          set(itemsHeight.value, id, node.h)
        }
        if (!gridStack || gridStack.getColumn() <= 1) {
          return
        }
        saveItemsLayout(items)
      })
    }

    function resizeItem(item: GridItem) {
      setTimeout(() => {
        if (!gridStack) {
          return
        }

        const el = Utils.getElement(`#gs-item-${item.id}`) as GridItemHTMLElement
        if (!el) {
          return
        }

        const node = el.gridstackNode
        if (!node || !node.sizeToContent) {
          return
        }

        gridStack.resizeToContent(el)
      }, 250)
    }

    return {
      rowItems,

      gsAutoPosition,
      gridStackEl,

      itemHeight,
      resizeItem,
    }
  },
})
</script>

<style lang="scss" scoped></style>
