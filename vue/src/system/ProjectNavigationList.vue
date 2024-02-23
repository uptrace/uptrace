<template>
  <v-list>
    <div v-for="(item, i) in items" :key="i">
      <v-list-item v-if="!item.children" :to="item.to">
        <v-list-item-action>
          <v-icon>{{ item.icon }}</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>{{ item.text }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-menu v-else open-on-hover offset-x right :nudge-left="1">
        <template #activator="{ on, attrs }">
          <v-list-item ripple :to="item.to" :title="item.text" v-bind="attrs" v-on="on">
            <v-list-item-action>
              <v-icon>{{ item.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
            <v-list-item-icon>
              <v-icon dense>mdi-chevron-right</v-icon>
            </v-list-item-icon>
          </v-list-item>
        </template>

        <v-list>
          <v-list-item v-for="child in item.children" :key="child.text" :to="child.to" exact-path>
            <v-list-item-title>{{ child.text }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </div>
  </v-list>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'ProjectNavigationList',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup(props) {
    const projectParams = computed(() => {
      return { projectId: props.project.id }
    })

    const items = computed(() => {
      const items: any[] = [
        {
          text: 'Overview',
          to: { name: 'Overview', params: projectParams.value },
          icon: 'mdi-monitor-dashboard',
          children: [
            {
              text: 'Systems',
              to: { name: 'SystemOverview', params: projectParams.value },
            },
            {
              text: 'Service graph',
              to: { name: 'ServiceGraph', params: projectParams.value },
            },
            {
              text: 'Slowest groups',
              to: { name: 'SlowestGroups', params: projectParams.value },
            },
          ],
        },
        {
          text: 'Traces & Logs',
          to: {
            name: 'SpanGroupList',
            params: projectParams.value,
          },
          icon: 'mdi-graph',
        },
        {
          text: 'Metrics',
          to: { name: 'DashboardList', params: projectParams.value },
          icon: 'mdi-chart-bar',
          children: [
            {
              text: 'Dashboards',
              to: { name: 'DashboardList', params: projectParams.value },
            },
            {
              text: 'Explore metrics',
              to: { name: 'MetricsExplore', params: projectParams.value },
            },
          ],
        },
        {
          text: 'Alerting',
          to: { name: 'AlertList', params: projectParams.value },
          icon: 'mdi-bell-outline',
          children: [
            {
              text: 'Alerts',
              to: { name: 'AlertList', params: projectParams.value },
            },
            {
              text: 'Monitors',
              to: { name: 'MonitorList', params: projectParams.value },
            },
            {
              text: 'Notifications channels',
              to: { name: 'NotifChannelList', params: projectParams.value },
            },
            {
              text: 'Email notifications',
              to: { name: 'NotifChannelEmail', params: projectParams.value },
            },
            {
              text: 'Annotations',
              to: { name: 'AnnotationList', params: projectParams.value },
            },
          ],
        },
      ]

      items.push({
        text: 'Project',
        to: { name: 'ProjectShow', params: projectParams.value },
        icon: 'mdi-cog-outline',
        children: [
          {
            text: 'Settings',
            to: { name: 'ProjectShow', params: projectParams.value },
          },
          {
            text: 'Data Source Name',
            to: { name: 'ProjectDsn', params: projectParams.value },
          },
        ],
      })

      return items
    })

    return {
      items,
    }
  },
})
</script>

<style lang="scss" scoped></style>
