<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col class="text-subtitle-1 text--primary">
          This notification channel allows to receive Uptrace alerts using AlertManager JSON v2 API.
          You can then use AlertManager to manage alerts and receive notifications.
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="8">
          <v-text-field
            v-model="channel.name"
            label="Channel name"
            hint="Short name that clearly describes the channel"
            persistent-hint
            outlined
            dense
            required
            :rules="rules.name"
            autofocus
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-text-field
            v-model="channel.params.url"
            label="AlertManager URL"
            placeholder="https://alertmanager-host.com:9093/api/v2/alerts"
            hint="Publicly accessible AlertManager endpoint"
            persistent-hint
            outlined
            dense
            :rules="rules.url"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-spacer />
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
import { useNotifChannelManager, WebhookNotifChannel } from '@/alerting/use-notif-channels'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'NotifChannelWebhookForm',

  props: {
    channel: {
      type: Object as PropType<WebhookNotifChannel>,
      required: true,
    },
  },

  setup(props, ctx) {
    const man = useNotifChannelManager()

    const form = shallowRef()
    const isValid = shallowRef(true)
    const rules = {
      name: [requiredRule],
      url: [requiredRule],
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
        return man.webhookUpdate(props.channel)
      }
      return man.webhookCreate(props.channel)
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
