<template>
  <v-container fluid class="bg--light">
    <v-row align="center" style="height: calc(100vh - 128px)">
      <v-col>
        <v-card max-width="600" class="mx-auto">
          <v-toolbar color="bg--primary" flat>
            <v-toolbar-title>Email notifications</v-toolbar-title>
          </v-toolbar>

          <v-container fluid>
            <v-row>
              <v-col class="pa-4">
                <v-skeleton-loader v-if="!email.channel" type="card"></v-skeleton-loader>

                <NotifChannelEmailForm v-else :channel="reactive(email.channel)" />
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, reactive } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useEmailChannel } from '@/alerting/use-notif-channels'

// Components
import NotifChannelEmailForm from '@/alerting/NotifChannelEmailForm.vue'

export default defineComponent({
  name: 'EmailNotifications',
  components: { NotifChannelEmailForm },

  setup() {
    useTitle('Email notifications')
    const email = useEmailChannel()

    return { email, reactive }
  },
})
</script>

<style lang="scss" scoped></style>
