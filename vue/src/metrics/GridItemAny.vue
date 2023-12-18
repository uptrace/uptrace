<template>
  <div>
    <GridItemSwitch
      v-bind="$props"
      v-on="$listeners"
      @click:edit="
        activeGridItem = reactive(cloneDeep(gridItem))
        dialog = true
      "
    />

    <v-dialog v-model="dialog" fullscreen>
      <GridItemFormSwitch
        v-if="activeGridItem"
        :date-range="dateRange"
        :dashboard="dashboard"
        :table-grouping="dashboard.tableGrouping"
        :grid-item="activeGridItem"
        @save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, reactive, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import GridItemSwitch from '@/metrics/GridItemSwitch.vue'
import GridItemFormSwitch from '@/metrics/GridItemFormSwitch.vue'

// Misc
import { Dashboard, GridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemAny',
  components: { GridItemSwitch, GridItemFormSwitch },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<GridItem>,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
  },

  setup(props) {
    const activeGridItem = shallowRef<GridItem>()
    const dialog = shallowRef(false)

    return {
      activeGridItem,
      dialog,

      cloneDeep,
      reactive,
    }
  },
})
</script>

<style lang="scss" scoped></style>
