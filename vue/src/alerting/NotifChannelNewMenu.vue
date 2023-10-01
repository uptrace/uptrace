<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on, attrs }">
      <v-btn color="primary" dark v-bind="attrs" v-on="on">
        Create channel
        <v-icon right>mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-list>
      <v-list-item @click="$emit('click:new', reactive(emptySlackNotifChannel()))">
        <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
        <v-list-item-title>Slack</v-list-item-title>
      </v-list-item>

      <v-list-item @click="$emit('click:new', reactive(emptyTelegramNotifChannel()))">
        <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
        <v-list-item-title>Telegram</v-list-item-title>
      </v-list-item>

      <v-list-item @click="$emit('click:new', reactive(emptyWebhookNotifChannel()))">
        <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
        <v-list-item-title>WebHook</v-list-item-title>
      </v-list-item>

      <v-list-item @click="$emit('click:new', reactive(emptyAlertmanagerNotifChannel()))">
        <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
        <v-list-item-title>AlertManager</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive } from 'vue'

// Composables
import {
  emptySlackNotifChannel,
  emptyTelegramNotifChannel,
  emptyWebhookNotifChannel,
  emptyAlertmanagerNotifChannel,
} from '@/alerting/use-notif-channels'

export default defineComponent({
  name: 'NotifChannelNewMenu',

  setup() {
    const menu = shallowRef(false)
    return {
      menu,
      emptySlackNotifChannel,
      emptyTelegramNotifChannel,
      emptyWebhookNotifChannel,
      emptyAlertmanagerNotifChannel,
      reactive,
    }
  },
})
</script>

<style lang="scss" scoped></style>
