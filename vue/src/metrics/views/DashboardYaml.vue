<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-row>
      <v-col cols="auto">
        <v-btn :loading="dashMan.pending" color="primary" @click="saveYaml">Save YAML</v-btn>
      </v-col>
      <v-col cols="auto">
        <v-btn :href="dashboard.yamlUrl" color="secondary">Download YAML</v-btn>
      </v-col>
    </v-row>

    <v-row v-if="!dash.status.hasData()">
      <v-col>
        <v-skeleton-loader type="image" loading></v-skeleton-loader>
      </v-col>
    </v-row>

    <v-row v-else>
      <v-col>
        <v-textarea
          v-model="yaml"
          auto-grow
          filled
          background-color="lime lighten-5"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="auto">
        <v-btn :loading="dashMan.pending" color="primary" @click="saveYaml">Save YAML</v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'
import { useYamlDashboard, useDashManager } from '@/metrics/use-dashboards'

// Utilities
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashYaml',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(_props, ctx) {
    const snackbar = useSnackbar()

    const yaml = shallowRef('')
    const dash = useYamlDashboard()
    const dashMan = useDashManager()

    watch(
      () => dash.yaml,
      (dashYaml) => {
        yaml.value = dashYaml
      },
      { immediate: true },
    )

    function saveYaml() {
      dashMan.updateYaml(yaml.value).then(() => {
        snackbar.notifySuccess(`The dashboard has beed successfully updated from the YAML`)
        ctx.emit('change')
      })
    }

    return { yaml, dash, dashMan, saveYaml }
  },
})
</script>

<style lang="scss" scoped></style>
