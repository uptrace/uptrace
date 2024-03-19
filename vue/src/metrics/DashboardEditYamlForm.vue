<template>
  <v-form v-model="isValid" @submit.prevent="saveYaml">
    <v-card>
      <v-toolbar color="bg--primary" flat>
        <v-toolbar-title>{{ dashboard.name }}</v-toolbar-title>

        <v-spacer />

        <v-btn color="primary" :disabled="!isValid" :loading="dashMan.pending" type="submit">
          Update dashboard
        </v-btn>
      </v-toolbar>

      <v-container>
        <v-row v-if="!dash.status.hasData()">
          <v-col>
            <v-skeleton-loader type="image" loading></v-skeleton-loader>
          </v-col>
        </v-row>

        <v-row v-else>
          <v-col>
            <v-sheet>
              <v-textarea
                v-model="yaml"
                rows="20"
                auto-grow
                filled
                hide-details="auto"
                :rules="rules.yaml"
                style="max-height: 75vh; overflow-y: auto"
              />
            </v-sheet>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useSnackbar } from '@/use/snackbar'
import { useYamlDashboard, useDashboardManager } from '@/metrics/use-dashboards'

// Misc
import { Dashboard } from '@/metrics/types'
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashboardEditYamlForm',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const isValid = shallowRef(true)
    const yaml = shallowRef('')
    const rules = {
      yaml: [requiredRule],
    }

    const snackbar = useSnackbar()
    const route = useRoute()
    const dashMan = useDashboardManager()
    const dash = useYamlDashboard(() => {
      if (!props.dashboard.id) {
        return
      }

      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/dashboards/${props.dashboard.id}/yaml`,
      }
    })

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
        ctx.emit('updated')
      })
    }

    return {
      isValid,
      rules,
      yaml,
      dash,
      dashMan,
      saveYaml,
    }
  },
})
</script>

<style lang="scss" scoped></style>
