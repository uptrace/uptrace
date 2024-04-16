<template>
  <v-form ref="form" v-model="isValid" lazy-validation class="bg--none" @submit.prevent="submit()">
    <v-card :max-width="maxWidth" class="mx-auto">
      <v-toolbar flat color="bg--light-primary">
        <slot name="title"></slot>

        <v-spacer />

        <v-col cols="auto">
          <v-btn v-if="monitor.id" text class="mr-2" @click="$emit('click:cancel', monitor)">
            Cancel
          </v-btn>

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
import { useProject } from '@/org/use-projects'
import { useMonitorManager } from '@/alerting/use-monitors'

// Misc
import { MetricMonitor, Monitor } from '@/alerting/types'

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

    const project = useProject()
    const monitorMan = useMonitorManager()
    function submit() {
      if (!form.value.validate()) {
        return
      }

      save().then((monitor: Monitor) => {
        ctx.emit('saved', monitor)
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

      project,
      monitorMan,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
