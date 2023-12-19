<template>
  <v-app>
    <AppBar
      v-if="header && user.isAuth && project.data"
      :user="user.current"
      :project="project.data"
    />

    <v-app-bar v-if="header" app absolute flat color="white" class="v-bar--underline">
      <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pa-0 fill-height">
        <v-row align="center" class="flex-nowrap">
          <v-col cols="auto">
            <div class="mt-2">
              <UptraceLogoSmall />
            </div>
          </v-col>

          <v-col cols="auto">
            <ProjectPicker />
          </v-col>

          <v-col v-if="!searchVisible" cols="auto">
            <v-tabs optional class="ml-lg-10 ml-xl-16">
              <template v-if="user.isAuth && $route.params.projectId">
                <v-tab :to="{ name: 'Overview' }">Overview</v-tab>
                <v-tab :to="{ name: 'SpanGroupList' }">Traces & Logs</v-tab>
                <v-tab :to="{ name: 'DashboardList' }">Dashboards</v-tab>
                <v-tab :to="{ name: 'Alerting' }">Alerts</v-tab>
              </template>
              <v-tab v-else-if="user.isAuth" :to="{ name: 'UserProfile' }">Profile</v-tab>
              <v-tab v-else :to="{ name: 'Login' }">Login</v-tab>
            </v-tabs>
          </v-col>

          <v-spacer />

          <v-col v-if="user.isAuth && $route.params.projectId" cols="auto">
            <AppSearch v-model="searchVisible" />
          </v-col>
          <v-col cols="auto">
            <v-menu v-if="user.isAuth" bottom offset-y>
              <template #activator="{ attrs, on }">
                <v-btn elevation="0" color="transparent" class="pl-1 pr-0" v-bind="attrs" v-on="on">
                  <v-avatar size="26px">
                    <img alt="Avatar" :src="user.current.avatar" />
                  </v-avatar>
                  <v-icon>mdi-menu-down</v-icon>
                </v-btn>
              </template>

              <v-list>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title class="font-weight-bold">
                      {{ user.current.name || 'Anonymous' }}
                    </v-list-item-title>
                    <v-list-item-subtitle v-if="user.current.email">
                      {{ user.current.email }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item :to="{ name: 'UserProfile' }">
                  <v-list-item-title>Profile</v-list-item-title>
                </v-list-item>

                <v-divider></v-divider>

                <v-list-item @click="user.logout">
                  <v-list-item-title>Sign out</v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-col>
        </v-row>
      </v-container>
    </v-app-bar>

    <v-main>
      <GlobalSnackbar />
      <GlobalConfirm />
      <router-view :date-range="dateRange" />
    </v-main>

    <v-footer v-if="footer" app absolute color="grey lighten-5">
      <v-container fluid>
        <v-row justify="center" align="center">
          <v-col cols="auto">
            <v-btn href="https://uptrace.dev/get/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-help-circle-outline</v-icon>
              <span>Documentation</span>
            </v-btn>
            <v-btn href="https://uptrace.dev/opentelemetry/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-open-source-initiative</v-icon>
              <span>OpenTelemetry</span>
            </v-btn>
            <v-btn href="https://uptrace.dev/get/instrument/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-toy-brick-outline</v-icon>
              <span>Instrumentations</span>
            </v-btn>
            <v-btn href="https://t.me/uptrace" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-message-outline</v-icon>
              <span>Telegram</span>
            </v-btn>
            <v-btn href="https://github.com/uptrace/uptrace" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-github</v-icon>
              <span>GitHub</span>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, shallowRef, provide } from 'vue'

// Composables
import { provideForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useUser } from '@/org/use-users'
import { useProject } from '@/org/use-projects'

// Components
import AppBar from '@/components/AppBar.vue'
import UptraceLogoSmall from '@/components/UptraceLogoSmall.vue'
import ProjectPicker from '@/components/ProjectPicker.vue'
import AppSearch from '@/components/AppSearch.vue'
import GlobalSnackbar from '@/components/GlobalSnackbar.vue'
import GlobalConfirm from '@/components/GlobalConfirm.vue'

export default defineComponent({
  name: 'App',
  components: {
    AppBar,
    UptraceLogoSmall,
    ProjectPicker,
    AppSearch,
    GlobalSnackbar,
    GlobalConfirm,
  },

  setup() {
    // Make these global across the app.
    provideForceReload()

    const header = shallowRef(true)
    provide('header', header)

    const footer = shallowRef(true)
    provide('footer', footer)

    const searchVisible = shallowRef(false)

    const dateRange = useDateRange()
    const user = useUser()
    const project = useProject()

    return {
      header,
      footer,
      searchVisible,

      dateRange,
      user,
      project,
    }
  },
})
</script>

<style lang="scss">
.theme--light,
.theme--dark {
  .v-bar--underline {
    border-width: 0 0 thin 0;
    border-style: solid;

    &.theme--light {
      border-bottom-color: #0000001f !important;
    }

    &.theme--dark {
      border-bottom-color: #ffffff1f !important;
    }
  }
}
</style>

<style lang="scss" scoped>
.v-footer strong {
  font-weight: 500;
}

.v-app-bar ::v-deep .v-toolbar__content {
  padding-top: 0px;
  padding-bottom: 0px;
}
</style>
