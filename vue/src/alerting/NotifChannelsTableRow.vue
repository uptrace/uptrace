<template>
  <tr class="cursor-pointer" @click="$router.push(channelRoute(channel))">
    <td>{{ channel.name }}</td>
    <td>{{ channel.type }}</td>
    <td class="text-center">
      <NotifChannelStateAvatar :state="channel.state" class="mr-2" />
    </td>
    <td class="text-center">
      <v-btn
        v-if="channel.state === NotifChannelState.Delivering"
        :loading="man.pending"
        icon
        title="Pause channel"
        @click.stop="pauseChannel(channel)"
      >
        <v-icon>mdi-pause</v-icon>
      </v-btn>
      <v-btn
        v-else
        :loading="man.pending"
        icon
        title="Unpause channel"
        @click.stop="unpauseChannel(channel)"
      >
        <v-icon>mdi-play</v-icon>
      </v-btn>

      <v-btn :loading="man.pending" icon title="Edit channel" :to="channelRoute(channel)">
        <v-icon>mdi-pencil-outline</v-icon>
      </v-btn>
      <v-btn
        :loading="man.pending"
        icon
        title="Delete channel"
        @click.stop="deleteChannel(channel)"
      >
        <v-icon>mdi-delete-outline</v-icon>
      </v-btn>
    </td>
  </tr>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import {
  useNotifChannelManager,
  NotifChannel,
  NotifChannelType,
  NotifChannelState,
} from '@/alerting/use-notif-channels'

// Components

import NotifChannelStateAvatar from '@/alerting/NotifChannelStateAvatar.vue'

export default defineComponent({
  name: 'NotifChannelsTable',
  components: { NotifChannelStateAvatar },

  props: {
    channel: {
      type: Object as PropType<NotifChannel>,
      required: true,
    },
  },

  setup(_props, ctx) {
    const confirm = useConfirm()
    const man = useNotifChannelManager()

    function pauseChannel(channel: NotifChannel) {
      man.pause(channel.id).then(() => {
        ctx.emit('change')
      })
    }

    function unpauseChannel(channel: NotifChannel) {
      man.unpause(channel.id).then(() => {
        ctx.emit('change')
      })
    }

    function deleteChannel(channel: NotifChannel) {
      confirm
        .open('Delete channel', `Do you really want to delete "${channel.name}" channel?`)
        .then(
          () => {
            man.delete(channel.id).then(() => {
              ctx.emit('change')
            })
          },
          () => {},
        )
    }

    function channelRoute(channel: NotifChannel) {
      switch (channel.type) {
        case NotifChannelType.Slack:
          return { name: 'NotifChannelShowSlack', params: { channelId: channel.id } }
        case NotifChannelType.Telegram:
          return { name: 'NotifChannelShowTelegram', params: { channelId: channel.id } }
        case NotifChannelType.Webhook:
          return { name: 'NotifChannelShowWebhook', params: { channelId: channel.id } }
        case NotifChannelType.Alertmanager:
          return { name: 'NotifChannelShowAlertmanager', params: { channelId: channel.id } }
      }
    }

    return {
      NotifChannelState,

      man,
      pauseChannel,
      unpauseChannel,
      deleteChannel,
      channelRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
