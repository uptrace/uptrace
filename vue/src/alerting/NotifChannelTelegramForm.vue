<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col>
          <p>To receive alerts via Telegram:</p>

          <ol class="mb-4">
            <li class="mb-1">
              Add <a href="https://t.me/UptraceBot" target="_blank">@UptraceBot</a> to the Telegram
              channel where you want to receive notifications.
            </li>
            <li>
              Get the channel's <code>chat.id</code> by adding
              <a href="https://t.me/RawDataBot" target="_blank">@RawDataBot</a> and running
              <code>/start</code>.
            </li>
          </ol>

          <p>Note that the chat id is not static and changes whenever you change permissions.</p>
        </v-col>
      </v-row>

      <v-row dense>
        <v-col>
          <v-text-field
            v-model="channel.name"
            label="Notification channel name"
            hint="Concise name that clearly describes the channel"
            outlined
            dense
            required
            :rules="rules.name"
            autofocus
            style="max-width: 400px"
          />
        </v-col>
      </v-row>

      <v-row dense>
        <v-col>
          <v-text-field
            v-model.number="channel.params.chatId"
            type="number"
            label="Telegram chat ID"
            placeholder="-1091960820965"
            outlined
            dense
            :rules="rules.chatId"
            style="max-width: 400px"
          />
        </v-col>
      </v-row>

      <v-row dense>
        <v-col cols="auto">
          <v-btn :loading="man.pending" type="submit" color="primary">{{
            channel.id ? 'Save' : 'Create'
          }}</v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useNotifChannelManager, TelegramNotifChannel } from '@/alerting/use-notif-channels'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'NotifChannelTelegramForm',

  props: {
    channel: {
      type: Object as PropType<TelegramNotifChannel>,
      required: true,
    },
  },

  setup(props, ctx) {
    const man = useNotifChannelManager()

    const form = shallowRef()
    const isValid = shallowRef(true)
    const rules = {
      name: [requiredRule],
      chatId: [requiredRule],
    }

    function submit() {
      save().then(() => {
        ctx.emit('click:save')
        ctx.emit('click:close')
      })
    }

    function save() {
      if (!form.value.validate()) {
        return Promise.reject()
      }

      if (props.channel.id) {
        return man.telegramUpdate(props.channel)
      }
      return man.telegramCreate(props.channel)
    }

    return {
      man,

      form,
      isValid,
      rules,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
