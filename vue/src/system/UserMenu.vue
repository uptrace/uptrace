<template>
  <v-menu open-on-hover offset-x right :nudge-left="1">
    <template #activator="{ attrs, on }">
      <v-list-item
        two-line
        :to="{ name: 'UserProfile' }"
        title="User"
        class="px-2"
        v-bind="attrs"
        v-on="on"
      >
        <v-list-item-avatar>
          <v-avatar size="36">
            <img alt="Avatar" :src="user.current.avatar" />
          </v-avatar>
        </v-list-item-avatar>
        <v-list-item-content>
          <v-list-item-title class="font-weight-bold">
            {{ user.current.name || user.current.email }}
          </v-list-item-title>
          <v-list-item-subtitle>{{ user.current.email }}</v-list-item-subtitle>
        </v-list-item-content>
        <v-list-item-icon>
          <v-icon dense>mdi-chevron-right</v-icon>
        </v-list-item-icon>
      </v-list-item>
    </template>

    <v-list>
      <v-list-item :to="{ name: 'UserProfile' }" exact-path>
        <v-list-item-title>Profile</v-list-item-title>
      </v-list-item>
      <v-list-item :to="{ name: 'DataUsage' }" exact-path>
        <v-list-item-title>Data usage</v-list-item-title>
      </v-list-item>
      <v-list-item @click="user.logout">
        <v-list-item-title>Sign out</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useUser } from '@/org/use-users'

export default defineComponent({
  name: 'UserMenu',

  setup() {
    const user = useUser()
    return { user }
  },
})
</script>

<style lang="scss" scoped></style>
