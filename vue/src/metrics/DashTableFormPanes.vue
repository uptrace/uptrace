<template>
  <v-form ref="form" v-model="isValid" lazy-validation class="bg--none" @submit.prevent="submit()">
    <v-card :max-width="maxWidth" class="mx-auto">
      <v-toolbar flat color="bg--light-primary">
        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>

        <v-toolbar-title> Edit table dashboard </v-toolbar-title>

        <v-col cols="auto">
          <slot name="actions">
            <ForceReloadBtn small />
          </slot>
        </v-col>

        <v-spacer />

        <v-col cols="auto">
          <v-btn :loading="dashMan.pending" color="primary" @click="submit()">Save</v-btn>
        </v-col>
      </v-toolbar>

      <splitpanes class="default-theme" style="height: calc(100vh - 64px)">
        <pane size="70">
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
        <pane size="30">
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
import { useDashboardManager } from '@/metrics/use-dashboards'

// Components
import ForceReloadBtn from '@/components/ForceReloadBtn.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashTableFormPanes',
  components: {
    Splitpanes,
    Pane,
    ForceReloadBtn,
  },

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    maxWidth: {
      type: String,
      default: '1700px',
    },
  },

  setup(props, ctx) {
    const form = shallowRef()
    const isValid = shallowRef(false)

    const dashMan = useDashboardManager()
    function submit() {
      if (!form.value.validate()) {
        return
      }

      dashMan.updateTable(props.dashboard).then((dash) => {
        ctx.emit('saved', dash)
      })
    }

    return {
      form,
      isValid,

      dashMan,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
