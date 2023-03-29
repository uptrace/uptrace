<template>
  <div>
    <PageToolbar :fluid="$vuetify.breakpoint.lgAndDown">
      <v-toolbar-title>Notification Channels</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.lgAndDown">
      <v-row>
        <v-col>
          <NotifChannelNewMenu
            @click:new="
              activeChannel = $event
              dialog = true
            "
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet rounded="lg" outlined class="mb-4">
            <div class="pa-4">
              <NotifChannelsTable
                :loading="channels.loading"
                :channels="channels.items"
                @change="channels.reload"
              />
            </div>
          </v-sheet>
        </v-col>
      </v-row>
    </v-container>

    <v-dialog v-model="dialog" max-width="700">
      <v-card v-if="activeChannel">
        <v-toolbar flat color="light-blue lighten-5">
          <v-toolbar-title>New {{ activeChannel.type }} notification channel</v-toolbar-title>
          <v-spacer />
          <v-toolbar-items>
            <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <NotifChannelSlackForm
          v-if="activeChannel.type === NotifChannelType.Slack"
          :channel="activeChannel"
          @click:save="channels.reload"
          @click:close="dialog = false"
        />
        <NotifChannelAlertmanagerForm
          v-else-if="activeChannel.type === NotifChannelType.Alertmanager"
          :channel="activeChannel"
          @click:save="channels.reload"
          @click:close="dialog = false"
        />
        <NotifChannelWebhookForm
          v-else-if="activeChannel.type === NotifChannelType.Webhook"
          :channel="activeChannel"
          @click:save="channels.reload"
          @click:close="dialog = false"
        />
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useForceReload } from '@/use/force-reload'
import { useNotifChannels, NotifChannel, NotifChannelType } from '@/alerting/use-notif-channels'

// Components
import ForceReloadBtn from '@/components/date/ForceReloadBtn.vue'
import NotifChannelNewMenu from '@/alerting/NotifChannelNewMenu.vue'
import NotifChannelSlackForm from '@/alerting/NotifChannelSlackForm.vue'
import NotifChannelWebhookForm from '@/alerting/NotifChannelWebhookForm.vue'
import NotifChannelAlertmanagerForm from '@/alerting/NotifChannelAlertmanagerForm.vue'
import NotifChannelsTable from '@/alerting/NotifChannelsTable.vue'

export default defineComponent({
  name: 'ChannelList',
  components: {
    ForceReloadBtn,
    NotifChannelNewMenu,
    NotifChannelSlackForm,
    NotifChannelWebhookForm,
    NotifChannelAlertmanagerForm,
    NotifChannelsTable,
  },

  setup() {
    useTitle('Notification Channels')
    const { forceReloadParams } = useForceReload()

    const channels = useNotifChannels(() => {
      return forceReloadParams.value
    })

    const internalDialog = shallowRef(false)
    const activeChannel = shallowRef<NotifChannel>()
    const dialog = computed({
      get(): boolean {
        return Boolean(internalDialog.value && activeChannel.value)
      },
      set(dialog: boolean) {
        internalDialog.value = dialog
        if (!dialog) {
          activeChannel.value = undefined
        }
      },
    })

    return {
      NotifChannelType,

      channels,

      dialog,
      activeChannel,
    }
  },
})
</script>

<style lang="scss" scoped></style>
