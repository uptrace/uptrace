<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn :loading="dashMan.pending" icon v-on="on">
        <v-icon>mdi-menu</v-icon>
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

      <v-divider />

      <v-dialog v-model="yamlDialog" max-width="800px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>View dashboard YAML</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardYamlCard
          v-if="yamlDialog"
          :dashboard="dashboard"
          @click:cancel="yamlDialog = false"
        />
      </v-dialog>

      <v-dialog v-model="editDialog" max-width="500px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>Edit dashboard</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardForm
          :dashboard="reactive(cloneDeep(dashboard))"
          @saved="
            closeDialog()
            $emit('updated', $event)
          "
          @click:cancel="editDialog = false"
        >
        </DashboardForm>
      </v-dialog>

      <v-list-item ripple @click="cloneDashboard()">
        <v-list-item-content>
          <v-list-item-title>Clone dashboard</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item ripple @click="resetDashboard()">
        <v-list-item-content>
          <v-list-item-title>
            {{ dashboard.templateId ? 'Reset dashboard from template' : 'Reset grid layout' }}
          </v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item ripple @click="deleteDashboard()">
        <v-list-item-content>
          <v-list-item-title>Delete dashboard</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, reactive, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import { useDashboardManager } from '@/metrics/use-dashboards'

// Components
import DashboardForm from '@/metrics/DashboardForm.vue'
import DashboardYamlForm from '@/metrics/DashboardYamlForm.vue'
import DashboardYamlCard from '@/metrics/DashboardYamlCard.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardMenu',
  components: { DashboardForm, DashboardYamlForm, DashboardYamlCard },

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const confirm = useConfirm()

    const menu = shallowRef(false)
    const newDialog = shallowRef(false)
    const newYamlDialog = shallowRef(false)
    const yamlDialog = shallowRef(false)
    const editDialog = shallowRef(false)
    function closeDialog() {
      editDialog.value = false
      newDialog.value = false
      newYamlDialog.value = false
      yamlDialog.value = false
      menu.value = false
    }

    const dashMan = useDashboardManager()

    function cloneDashboard() {
      dashMan.clone(props.dashboard).then((dash) => {
        ctx.emit('created', dash)
      })
    }

    function resetDashboard() {
      dashMan.reset(props.dashboard).then(() => {
        ctx.emit('updated')
      })
    }

    function deleteDashboard() {
      confirm
        .open('Delete', `Do you really want to delete "${props.dashboard.name}" dashboard?`)
        .then(() => {
          dashMan.delete(props.dashboard).then(() => {
            ctx.emit('deleted')
          })
        })
        .catch(() => {})
    }

    return {
      menu,
      newDialog,
      newYamlDialog,
      yamlDialog,
      editDialog,
      closeDialog,

      dashMan,
      cloneDashboard,
      resetDashboard,
      deleteDashboard,

      cloneDeep,
      reactive,
    }
  },
})
</script>

<style lang="scss" scoped></style>
