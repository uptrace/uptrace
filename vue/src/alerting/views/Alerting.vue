<template>
  <div>
    <div class="border-bottom">
      <v-container :fluid="!$vuetify.breakpoint.xlOnly" class="py-0">
        <v-toolbar flat color="transparent" height="auto">
          <v-breadcrumbs large :items="breadcrumbs" divider=">" class="px-0"></v-breadcrumbs>

          <template #extension>
            <v-tabs :key="$route.fullPath" background-color="transparent">
              <v-tab :to="{ name: 'AlertList' }">Alerts</v-tab>
              <v-tab :to="{ name: 'MonitorList' }">Monitors</v-tab>
              <v-tab :to="{ name: 'NotifChannelList' }">Channels</v-tab>
            </v-tabs>
          </template>
        </v-toolbar>
      </v-container>
    </div>

    <router-view name="alerting" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useUser } from '@/org/use-users'
import { useProject } from '@/org/use-projects'

export default defineComponent({
  name: 'Alerting',

  setup() {
    const user = useUser()
    const project = useProject()

    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: project.data?.name ?? 'Project',
        to: {
          name: 'ProjectShow',
        },
      })

      bs.push({ text: 'Alerting' })

      return bs
    })

    return { user, breadcrumbs }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
