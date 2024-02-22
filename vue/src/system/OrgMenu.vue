<template>
  <v-menu open-on-hover offset-x right :nudge-left="1">
    <template #activator="{ attrs, on }">
      <v-list-item ripple v-bind="attrs" v-on="on">
        <v-list-item-action>
          <v-icon>mdi-account-group-outline</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>Organization</v-list-item-title>
        </v-list-item-content>
        <v-list-item-icon>
          <v-icon dense>mdi-chevron-right</v-icon>
        </v-list-item-icon>
      </v-list-item>
    </template>

    <v-list>
      <v-list-item :to="{ name: 'OrgShow', params: { orgId: project.orgId } }">
        <v-list-item-content>
          <v-list-item-title>Projects</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item :to="{ name: 'OrgMembers', params: { orgId: project.orgId } }">
        <v-list-item-content>
          <v-list-item-title>Teams and members</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item :to="{ name: 'UserInvite', params: { orgId: project.orgId } }">
        <v-list-item-content>
          <v-list-item-title>Invite users</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-list-item :to="{ name: 'OrgBilling', params: { orgId: project.orgId } }">
        <v-list-item-content>
          <v-list-item-title>Billing and usage</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'OrgMenu',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
