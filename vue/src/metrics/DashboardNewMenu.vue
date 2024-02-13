<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn color="primary" v-on="on">
        New dashboard
        <v-icon right>mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <v-list>
      <v-dialog v-model="newDialog" max-width="500px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>New dashboard</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardForm
          @saved="
            closeDialog()
            $emit('created', $event)
          "
          @click:cancel="newDialog = false"
        >
        </DashboardForm>
      </v-dialog>

      <v-dialog v-model="newYamlDialog" max-width="800px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>New dashboard from YAML</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardYamlForm
          v-if="newYamlDialog"
          @created="
            closeDialog()
            $emit('created', $event)
          "
          @click:cancel="newYamlDialog = false"
        />
      </v-dialog>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Components
import DashboardForm from '@/metrics/DashboardForm.vue'
import DashboardYamlForm from '@/metrics/DashboardYamlForm.vue'

export default defineComponent({
  name: 'DashboardMenu',
  components: { DashboardForm, DashboardYamlForm },

  setup() {
    const menu = shallowRef(false)
    const newDialog = shallowRef(false)
    const newYamlDialog = shallowRef(false)

    function closeDialog() {
      newDialog.value = false
      newYamlDialog.value = false
      menu.value = false
    }

    return {
      menu,
      newDialog,
      newYamlDialog,
      closeDialog,
    }
  },
})
</script>

<style lang="scss" scoped></style>
