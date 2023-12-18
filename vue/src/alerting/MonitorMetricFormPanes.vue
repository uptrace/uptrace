<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit()">
    <v-card :max-width="maxWidth" class="mx-auto">
      <v-toolbar flat color="light-blue lighten-5">
        <slot name="title"></slot>

        <v-spacer />

        <v-col cols="auto">
          <v-btn :loading="monitorMan.pending" color="primary" @click="submit()">
            {{ monitor.id ? 'Update' : 'Create' }}
          </v-btn>
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
              <v-container fluid class="mx-auto fill-height">
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
          <slot name="options" :form="form" />
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
import { useMonitorManager } from '@/alerting/use-monitors'

// Misc
import { MetricMonitor } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorMetricFormPanes',
  components: {
    Splitpanes,
    Pane,
  },

  props: {
    monitor: {
      type: Object as PropType<MetricMonitor>,
      required: true,
    },
    maxWidth: {
      type: Number,
      default: 1416,
    },
  },

  setup(props, ctx) {
    const form = shallowRef()
    const isValid = shallowRef(false)

    const monitorMan = useMonitorManager()
    function submit() {
      if (!form.value.validate()) {
        return
      }

      save().then(() => {
        ctx.emit('saved')
      })
    }
    function save() {
      if (props.monitor.id) {
        return monitorMan.updateMetricMonitor(props.monitor)
      }
      return monitorMan.createMetricMonitor(props.monitor)
    }

    return {
      form,
      isValid,

      monitorMan,
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
