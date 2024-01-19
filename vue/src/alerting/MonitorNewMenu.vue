<template>
  <div>
    <v-menu v-model="menu" offset-y>
      <template #activator="{ on, attrs }">
        <v-btn color="primary" dark v-bind="attrs" v-on="on">
          Create monitor
          <v-icon right>mdi-menu-down</v-icon>
        </v-btn>
      </template>
      <v-list>
        <v-list-item :to="{ name: 'MonitorMetricNew' }">
          <v-list-item-icon><v-icon>mdi-chart-line</v-icon></v-list-item-icon>
          <v-list-item-title>Create metrics monitor</v-list-item-title>
        </v-list-item>
        <v-list-item :to="{ name: 'MonitorErrorNew' }">
          <v-list-item-icon><v-icon>mdi-bug-outline</v-icon></v-list-item-icon>
          <v-list-item-title>Create errors monitor</v-list-item-title>
        </v-list-item>
        <v-list-item @click="dialog = true">
          <v-list-item-icon><v-icon>mdi-code-string</v-icon></v-list-item-icon>
          <v-list-item-title>New monitor from YAML</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>

    <DialogCard v-model="dialog" max-width="800" title="New monitor from YAML">
      <MonitorYamlForm
        v-if="dialog"
        @create="
          $emit('create')
          dialog = false
        "
      />
    </DialogCard>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Components
import DialogCard from '@/components/DialogCard.vue'
import MonitorYamlForm from '@/alerting/MonitorYamlForm.vue'

export default defineComponent({
  name: 'MonitorNewMenu',

  components: { DialogCard, MonitorYamlForm },

  setup() {
    const dialog = shallowRef(false)
    const menu = shallowRef(false)

    return { dialog, menu }
  },
})
</script>

<style lang="scss" scoped></style>
