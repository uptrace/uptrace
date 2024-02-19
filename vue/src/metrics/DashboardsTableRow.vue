<template>
  <tr
    class="cursor-pointer"
    @click="$router.push({ name: 'DashboardShow', params: { dashId: dashboard.id } })"
  >
    <td>
      <DashPinBtn :dashboard="dashboard" @update="$emit('change')" />
      <router-link
        :to="{ name: 'DashboardShow', params: { dashId: dashboard.id } }"
        class="text--primary text-decoration-none"
        @click.native.stop
        >{{ dashboard.name }}</router-link
      >
    </td>
    <td>
      <DateValue v-if="dashboard.updatedAt" :value="dashboard.updatedAt" format="relative" />
    </td>
    <td class="text-right text-no-wrap">
      <v-btn
        icon
        title="Delete dashboard"
        :loading="dashboardMan.pending"
        @click.stop="deleteDashboard"
      >
        <v-icon>mdi-delete-outline</v-icon>
      </v-btn>
      <v-menu v-model="menu" offset-y>
        <template #activator="{ on: onMenu, attrs }">
          <v-btn icon v-bind="attrs" v-on="onMenu">
            <v-icon>mdi-dots-vertical</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="openDashboardYamlDialog(dashboard)">
            <v-list-item-content>
              <v-list-item-title>View YAML</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item @click="cloneDashboard">
            <v-list-item-content>
              <v-list-item-title>Clone</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
    </td>

    <v-dialog v-if="yamlDialog" v-model="yamlDialog" max-width="800px">
      <DashboardYamlCard
        v-if="yamlDialog"
        :dashboard="dashboard"
        @click:cancel="yamlDialog = false"
      />
    </v-dialog>

    <v-dialog v-if="editDialog" v-model="editDialog" max-width="500" :title="dashboard.name">
      <DashboardForm
        :dashboard="reactive(cloneDeep(dashboard))"
        @saved="
          editDialog = false
          $emit('change', $event)
        "
        @click:cancel="editDialog = false"
      >
      </DashboardForm>
    </v-dialog>
  </tr>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, reactive, shallowRef, PropType } from 'vue'

// Composables
import { useRouterOnly } from '@/use/router'
import { useConfirm } from '@/use/confirm'
import { useDashboardManager } from '@/metrics/use-dashboards'

// Components
import DashboardForm from '@/metrics/DashboardForm.vue'
import DashboardYamlCard from '@/metrics/DashboardYamlCard.vue'
import DashPinBtn from '@/metrics/DashPinBtn.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardsTableRow',
  components: { DashboardForm, DashboardYamlCard, DashPinBtn },

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    headers: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const router = useRouterOnly()
    const menu = shallowRef(false)
    const activeDashboard = shallowRef<Dashboard>()
    const editDialog = shallowRef(false)
    const yamlDialog = shallowRef(false)
    const confirm = useConfirm()
    const dashboardMan = useDashboardManager()

    function openDashboardYamlDialog(dashboard: Dashboard) {
      activeDashboard.value = dashboard
      yamlDialog.value = true
    }

    function cloneDashboard() {
      dashboardMan.clone(props.dashboard).then((dash) => {
        ctx.emit('change', dash)

        router.push({
          name: 'DashboardTable',
          params: { dashId: String(dash.id) },
        })
      })
    }

    function deleteDashboard() {
      confirm
        .open('Delete', `Do you really want to delete "${props.dashboard.name}" dashboard?`)
        .then(() => {
          dashboardMan.delete(props.dashboard).then(() => {
            ctx.emit('change')
          })
        })
        .catch(() => {})
    }

    return {
      cloneDeep,
      reactive,

      menu,
      yamlDialog,
      editDialog,
      activeDashboard,

      dashboardMan,
      cloneDashboard,
      deleteDashboard,
      openDashboardYamlDialog,
    }
  },
})
</script>

<style lang="scss" scoped></style>
