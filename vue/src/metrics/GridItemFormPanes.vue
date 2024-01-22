<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit()">
    <v-card :max-width="maxWidth" class="mx-auto">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>

        <v-toolbar-title>
          {{ gridItem.id ? `Edit ${gridItem.type} grid item` : `New ${gridItem.type} grid item` }}
        </v-toolbar-title>

        <v-col cols="auto">
          <slot name="actions">
            <ForceReloadBtn small />
          </slot>
        </v-col>

        <v-spacer />

        <v-col cols="auto">
          <v-btn :loading="gridItemMan.pending" color="primary" @click="submit()">Save</v-btn>
        </v-col>
      </v-toolbar>

      <splitpanes class="default-theme" style="height: calc(100vh - 64px)">
        <pane size="65">
          <splitpanes horizontal>
            <pane size="30" min-size="10">
              <v-container fluid class="mx-auto fill-height" style="max-width: 920px">
                <v-row align="center">
                  <v-col>
                    <slot name="picker" />
                  </v-col>
                </v-row>
              </v-container>
            </pane>
            <pane size="70">
              <v-container fluid class="mx-auto fill-height" style="max-width: 920px">
                <v-row align="center">
                  <v-col>
                    <slot name="preview" />
                  </v-col>
                </v-row>
              </v-container>
            </pane>
          </splitpanes>
        </pane>
        <pane size="35">
          <slot name="options" />
        </pane>
      </splitpanes>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'

import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'
import { useGridItemManager } from '@/metrics/use-dashboards'

// Components
import ForceReloadBtn from '@/components/ForceReloadBtn.vue'

// Misc
import { ChartGridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemFormPanes',
  components: {
    Splitpanes,
    Pane,
    ForceReloadBtn,
  },

  props: {
    gridItem: {
      type: Object as PropType<ChartGridItem>,
      required: true,
    },
    maxWidth: {
      type: String,
      default: '1400px',
    },
  },

  setup(props, ctx) {
    const snackbar = useSnackbar()

    const form = shallowRef()
    const isValid = shallowRef(false)

    const gridItemMan = useGridItemManager()
    function submit() {
      if (!form.value.validate()) {
        return
      }
      if ('metrics' in props.gridItem.params && !props.gridItem.params.metrics.length) {
        snackbar.notifyError(`Please select a metric and click the "Apply" button`)
        return
      }
      gridItemMan.save(props.gridItem).then((gridItem) => {
        ctx.emit('save', gridItem)
      })
    }

    return {
      form,
      isValid,

      gridItemMan,
      submit,
    }
  },
})
</script>

<style lang="scss">
.splitpanes.default-theme {
  .splitpanes__pane {
    background-color: #fff;
  }
  .splitpanes__splitter {
    background-color: #f2f2f2;
  }
}
</style>

<style lang="scss" scoped>
.v-form {
  background-color: #f2f2f2;
}

.splitpanes__pane {
  overflow-y: auto;
}
</style>
