<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn :loading="dashMan.pending" icon v-on="on">
        <v-icon>mdi-dots-vertical</v-icon>
      </v-btn>
    </template>

    <v-list>
      <v-dialog v-model="editDialog" max-width="500px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>Settings</v-list-item-title>
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

      <v-list-item ripple @click="resetDashboard()">
        <v-list-item-content>
          <v-list-item-title>
            {{ dashboard.templateId ? 'Reset from template' : 'Reset grid layout' }}
          </v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-dialog v-model="yamlDialog" max-width="800px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>View YAML</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardYamlCard
          v-if="yamlDialog"
          :dashboard="dashboard"
          @click:cancel="yamlDialog = false"
        />
      </v-dialog>

      <v-dialog v-model="editYamlDialog" max-width="800px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-content>
              <v-list-item-title>Edit YAML</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashboardEditYamlForm
          v-if="editYamlDialog"
          :dashboard="dashboard"
          @updated="
            editYamlDialog = false
            $emit('updated', $event)
          "
          @click:cancel="editYamlDialog = false"
        />
      </v-dialog>

      <v-list-item ripple @click="cloneDashboard()">
        <v-list-item-content>
          <v-list-item-title>Clone</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item ripple @click="deleteDashboard()">
        <v-list-item-content>
          <v-list-item-title>Delete</v-list-item-title>
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
import DashboardYamlCard from '@/metrics/DashboardYamlCard.vue'
import DashboardEditYamlForm from '@/metrics/DashboardEditYamlForm.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardMenu',
  components: { DashboardForm, DashboardYamlCard, DashboardEditYamlForm },

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const confirm = useConfirm()

    const menu = shallowRef(false)
    const yamlDialog = shallowRef(false)
    const editYamlDialog = shallowRef(false)
    const editDialog = shallowRef(false)
    function closeDialog() {
      editDialog.value = false
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
      yamlDialog,
      editYamlDialog,
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
