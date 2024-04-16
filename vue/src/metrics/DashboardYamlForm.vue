<template>
  <v-form v-model="isValid" @submit.prevent="saveYaml">
    <v-card>
      <v-toolbar color="bg--light-primary" flat>
        <v-toolbar-title>New dashboard from YAML</v-toolbar-title>

        <v-spacer />

        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid>
        <v-row>
          <v-col>
            <v-textarea
              v-model="yaml"
              placeholder="To create a dashboard, paste here a dashboard template in YAML format"
              rows="20"
              auto-grow
              filled
              hide-details="auto"
              :rules="rules.yaml"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col class="text-right">
            <v-btn
              :disabled="!isValid"
              :loading="dashboardMan.pending"
              color="primary"
              type="submit"
            >
              Create dashboard
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useDashboardManager } from '@/metrics/use-dashboards'

// Misc
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashboardYamlForm',

  setup(_props, ctx) {
    const isValid = shallowRef(true)
    const rules = {
      yaml: [requiredRule],
    }

    const yaml = shallowRef('')
    const dashboardMan = useDashboardManager()

    function saveYaml() {
      dashboardMan.createYaml(yaml.value).then((dashboard) => {
        ctx.emit('created', dashboard)
      })
    }

    return {
      isValid,
      rules,
      yaml,
      dashboardMan,

      saveYaml,
    }
  },
})
</script>

<style lang="scss" scoped></style>
