<template>
  <v-card flat :color="expanded && gridRow.items.length ? 'transparent' : 'bg--foreground'">
    <v-hover v-slot="{ hover }">
      <div class="d-flex align-center" style="height: 32px">
        <div class="text-subtitle-1 font-weight-medium cursor-pointer" @click="toggle()">
          <v-icon class="mr-1">{{ expanded ? 'mdi-chevron-down' : 'mdi-chevron-right' }}</v-icon>
          <span>{{ row.title }}</span>
        </div>

        <div v-show="hover" class="ml-3">
          <v-btn icon title="Edit the row" @click="openDialog()">
            <v-icon>mdi-pencil-outline</v-icon>
          </v-btn>
          <v-btn :loading="gridRowMan.pending" icon title="Delete the row" @click="deleteRow()">
            <v-icon>mdi-delete-outline</v-icon>
          </v-btn>
        </div>

        <div v-show="hover" class="ml-3">
          <v-btn
            :loading="gridRowMan.pending"
            :disabled="row.index === 0"
            icon
            title="Move the row up"
            @click="moveUp()"
          >
            <v-icon>mdi-arrow-up</v-icon>
          </v-btn>
          <v-btn :loading="gridRowMan.pending" icon title="Move the row down" @click="moveDown()">
            <v-icon>mdi-arrow-down</v-icon>
          </v-btn>
        </div>
      </div>
    </v-hover>

    <template v-if="expanded">
      <GridStackCard :items="gridItems" :row-id="row.id" :row-index="row.index">
        <template #item="{ attrs, on }">
          <slot name="item" v-bind="{ attrs, on: { ...on, change: gridRow.reload } }" />
        </template>
      </GridStackCard>
    </template>

    <DialogCard v-model="dialog" max-width="800" title="Edit grid row">
      <GridRowForm
        :row="reactive(cloneDeep(row))"
        @save="
          $emit('change')
          dialog = false
        "
        @click:cancel="dialog = false"
      />
    </DialogCard>
  </v-card>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, reactive, computed, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { injectForceReload } from '@/use/force-reload'
import { useConfirm } from '@/use/confirm'
import { useGridRow, useGridRowManager } from '@/metrics/use-dashboards'

// Components
import DialogCard from '@/components/DialogCard.vue'
import GridStackCard from '@/metrics/GridStackCard.vue'
import GridRowForm from '@/metrics/GridRowForm.vue'

// Misc
import { GridRow } from '@/metrics/types'

export default defineComponent({
  name: 'GridRowCard',
  components: { DialogCard, GridStackCard, GridRowForm },

  props: {
    row: {
      type: Object as PropType<GridRow>,
      required: true,
    },
  },

  setup(props, ctx) {
    const route = useRoute()

    const expanded = shallowRef(false)
    function toggle() {
      expanded.value = !expanded.value
    }
    watch(
      () => props.row.expanded,
      (expandedValue) => {
        expanded.value = expandedValue
      },
      { immediate: true },
    )

    const forceReload = injectForceReload()
    const gridRow = useGridRow(() => {
      if (!expanded.value) {
        return null
      }

      const { projectId, dashId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/dashboards/${dashId}/rows/${props.row.id}`,
        params: {
          ...forceReload.params,
        },
      }
    })
    watch(
      () => props.row,
      () => {
        gridRow.reload()
      },
      { flush: 'sync' },
    )

    const gridItems = computed(() => {
      if (gridRow.status.hasData()) {
        return gridRow.items
      }
      return props.row.items
    })

    const confirm = useConfirm()
    const gridRowMan = useGridRowManager()
    function deleteRow() {
      confirm.open('Delete', `Do you really want to delete the "${props.row.title}" row?`).then(
        () => {
          gridRowMan.delete(props.row).then(() => {
            ctx.emit('change')
          })
        },
        () => {},
      )
    }
    function moveUp() {
      gridRowMan.moveUp(props.row).then(() => {
        ctx.emit('change')
      })
    }
    function moveDown() {
      gridRowMan.moveDown(props.row).then(() => {
        ctx.emit('change')
      })
    }

    const dialog = shallowRef(false)
    function openDialog() {
      dialog.value = true
    }

    return {
      gridRow,
      gridItems,

      expanded,
      toggle,

      dialog,
      openDialog,

      gridRowMan,
      deleteRow,
      moveUp,
      moveDown,

      cloneDeep,
      reactive,
    }
  },
})
</script>

<style lang="scss" scoped></style>
