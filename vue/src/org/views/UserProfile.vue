<template>
  <div class="container--fixed-sm">
    <portal to="navigation">
      <v-tabs :key="$route.fullPath" background-color="transparent">
        <v-tab :to="{ name: 'UserProfile' }" exact-path> Profile </v-tab>
        <v-tab :to="{ name: 'DataUsage' }" exact-path> Data usage </v-tab>
      </v-tabs>
    </portal>

    <div v-if="user.isAuth">
      <PageToolbar>
        <v-toolbar-title>{{ title }}</v-toolbar-title>
      </PageToolbar>

      <v-container class="py-6">
        <v-row>
          <v-col>
            You can change user settings in the <code>uptrace.yml</code> config file. See
            <a href="https://uptrace.dev/get/config.html#managing-users" target="_blank"
              >documentation</a
            >
            for details.
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="6">
            <v-text-field v-model="user.current.name" disabled label="Name" filled required />

            <v-text-field v-model="user.current.email" disabled label="Email" filled required />

            <v-checkbox
              v-model="user.current.notifyByEmail"
              disabled
              label="Allow to send alert notifications via email"
              hide-details="auto"
              class="mt-0"
            />
          </v-col>
        </v-row>

        <v-row v-if="user.current.authToken">
          <v-col>
            <div class="text-body-2 text--secondary">Auth token</div>
            <PrismCode :code="user.current.authToken" />
          </v-col>
        </v-row>
      </v-container>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import { useTitle } from '@vueuse/core'
import { useUser } from '@/org/use-users'

export default defineComponent({
  name: 'UserProfile',

  setup() {
    const title = 'Profile'
    useTitle(title)

    const user = useUser()

    return {
      title,
      user,
    }
  },
})
</script>

<style lang="scss" scoped></style>
