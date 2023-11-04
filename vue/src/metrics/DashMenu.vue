<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn :loading="dashMan.pending" icon v-on="on">
        <v-icon>mdi-menu</v-icon>
      </v-btn>
    </template>

    <v-list>
      <v-dialog v-if="dashboard.data" v-model="newDialog" max-width="500px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-action>
              <v-icon>mdi-playlist-plus</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>New dashboard</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashNewForm @create="onCreateDash">
          <template #prepend-actions>
            <v-btn color="blue darken-1" text @click="closeDialog">Cancel</v-btn>
          </template>
        </DashNewForm>
      </v-dialog>

      <v-dialog v-if="dashboard.data" v-model="editDialog" max-width="500px">
        <template #activator="{ on }">
          <v-list-item ripple v-on="on">
            <v-list-item-action>
              <v-icon>mdi-playlist-edit</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>Edit dashboard</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>

        <DashEditForm
          :dashboard="dashboard.data"
          @update="onUpdateDash"
          @click:cancel="closeDialog"
        >
        </DashEditForm>
      </v-dialog>

      <v-list-item ripple @click="cloneDash">
        <v-list-item-action>
          <v-icon>mdi-playlist-play</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>Clone dashboard</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item ripple @click="deleteDash">
        <v-list-item-action>
          <v-icon>mdi-playlist-minus</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>Delete dashboard</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useDashManager, UseDashboards, UseDashboard } from '@/metrics/use-dashboards'

// Components
import DashNewForm from '@/metrics/DashNewForm.vue'
import DashEditForm from '@/metrics/DashEditForm.vue'

// Types
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashMenu',
  components: {
    DashNewForm,
    DashEditForm,
  },

  props: {
    dashboards: {
      type: Object as PropType<UseDashboards>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
  },

  setup(props) {
    const { router } = useRouter()
    const menu = shallowRef(false)
    const newDialog = shallowRef(false)
    const editDialog = shallowRef(false)

    const dashMan = useDashManager()

    function onUpdateDash() {
      editDialog.value = false
      menu.value = false
      props.dashboards.reload()
      props.dashboard.reload()
    }

    function onCreateDash(dash: Dashboard) {
      newDialog.value = false
      menu.value = false
      props.dashboards.reload().then(() => {
        router.push({ name: 'MetricsDashShow', params: { dashId: String(dash.id) } })
      })
    }

    function closeDialog() {
      editDialog.value = false
      newDialog.value = false
      menu.value = false
    }

    function cloneDash() {
      if (!props.dashboard.data) {
        return
      }

      dashMan.clone(props.dashboard.data).then((dash) => {
        props.dashboards.reload().then(() => {
          router.push({ name: 'MetricsDashShow', params: { dashId: String(dash.id) } })
        })
      })
    }

    function deleteDash() {
      if (!props.dashboard.data) {
        return
      }

      dashMan.delete(props.dashboard.data).then(() => {
        props.dashboards.reload()
      })
    }

    return {
      menu,
      newDialog,
      editDialog,

      dashMan,

      onUpdateDash,
      onCreateDash,
      closeDialog,

      cloneDash,
      deleteDash,
    }
  },
})
</script>

<style lang="scss" scoped></style>
