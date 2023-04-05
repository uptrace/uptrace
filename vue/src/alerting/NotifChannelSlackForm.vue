<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col class="text-subtitle-1 text--primary">
          <p>
            To receive notifications via Slack, you need to create an
            <a href="https://api.slack.com/messaging/webhooks" target="_blank">incoming webhook</a>
            on Slack and use the created webhook URL to configure Uptrace.
          </p>
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
            v-model="channel.params.webhookUrl"
            label="Slack webhook URL"
            placeholder="https://hooks.slack.com/services/********"
            outlined
            dense
            hide-details="auto"
            :rules="rules.webhookUrl"
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
import { useNotifChannelManager, SlackNotifChannel } from '@/alerting/use-notif-channels'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'NotifChannelSlackForm',

  props: {
    channel: {
      type: Object as PropType<SlackNotifChannel>,
      required: true,
    },
  },

  setup(props, ctx) {
    const man = useNotifChannelManager()

    const form = shallowRef()
    const isValid = shallowRef(true)
    const rules = {
      name: [requiredRule],
      webhookUrl: [requiredRule],
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
        return man.slackUpdate(props.channel)
      }
      return man.slackCreate(props.channel)
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
