<template>
  <div>
    <portal to="navigation">
      <v-tabs :key="$route.fullPath" background-color="transparent">
        <v-tab :to="{ name: 'AlertList' }">Alerts</v-tab>
        <v-tab :to="{ name: 'MonitorList' }">Monitors</v-tab>
        <v-tab :to="{ name: 'NotifChannelList' }">Channels</v-tab>
        <v-tab :to="{ name: 'NotifChannelEmail' }">Email notifications</v-tab>
        <v-tab :to="{ name: 'AnnotationList' }">Annotations</v-tab>
      </v-tabs>
    </portal>

    <router-view name="alerting">
      <template #breadcrumbs>
        <v-breadcrumbs large :items="breadcrumbs" divider=">" class="px-1"></v-breadcrumbs>
      </template>
    </router-view>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useProject } from '@/org/use-projects'

export default defineComponent({
  name: 'Alerting',

  setup() {
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

    return { breadcrumbs }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
