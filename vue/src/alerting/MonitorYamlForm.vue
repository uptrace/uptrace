<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-form v-model="isValid" @submit.prevent="saveYaml">
      <v-row>
        <v-col>
          <v-textarea
            v-model="yaml"
            placeholder="To create a monitor, paste here a a monitor definition in YAML format"
            auto-grow
            filled
            hide-details="auto"
            :rules="rules.yaml"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col class="text-right">
          <v-btn :disabled="!isValid" :loading="monitorMan.pending" color="primary" type="submit">
            Create monitor
          </v-btn>
        </v-col>
      </v-row>
    </v-form>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useMonitorManager, routeForMonitor } from '@/alerting/use-monitors'
import { useRouterOnly } from '@/use/router'

// Misc
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'MonitorYamlForm',

  setup(_props, ctx) {
    const isValid = shallowRef(true)
    const rules = {
      yaml: [requiredRule],
    }

    const router = useRouterOnly()
    const yaml = shallowRef('')
    const monitorMan = useMonitorManager()

    function saveYaml() {
      monitorMan.createMonitorFromYaml(yaml.value).then((monitors) => {
        ctx.emit('create')

        if (monitors.length === 1) {
          router.push(routeForMonitor(monitors[0]))
        }
      })
    }

    return {
      isValid,
      rules,
      yaml,
      monitorMan,

      saveYaml,
    }
  },
})
</script>

<style lang="scss" scoped></style>
