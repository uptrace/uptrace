<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col class="text-subtitle-1 text--primary">
          <p>
            Slack notifications can be configured using either webhooks or bot tokens. Choose the
            method that best fits your setup:
          </p>
          <ul>
            <li><strong>Webhook:</strong> Simple setup using incoming webhooks</li>
            <li><strong>Bot Token:</strong> More flexible with richer API access</li>
          </ul>
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
        <v-col cols="12">
          <v-btn-toggle
            v-model="channel.params.authMethod"
            mandatory
            color="primary"
            group
            class="v-btn-group--horizontal"
          >
            <v-btn value="webhook" class="text-none">
              <v-icon left>mdi-webhook</v-icon>
              Webhook
            </v-btn>
            <v-btn value="token" class="text-none">
              <v-icon left>mdi-robot</v-icon>
              Bot Token
            </v-btn>
          </v-btn-toggle>
        </v-col>
      </v-row>

      <!-- Webhook Configuration -->
      <template v-if="channel.params.authMethod === 'webhook' || !channel.params.authMethod">
        <v-row>
          <v-col class="text-subtitle-2 text--primary">
            <p>
              Create an
              <a href="https://api.slack.com/messaging/webhooks" target="_blank"
                >incoming webhook</a
              >
              on Slack and paste the URL below.
            </p>
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
      </template>

      <!-- Token Configuration -->
      <template v-if="channel.params.authMethod === 'token'">
        <v-row>
          <v-col class="text-subtitle-2 text--primary">
            <p>
              Create a
              <a href="https://api.slack.com/apps" target="_blank">Slack app</a>
              and generate a bot token with the <code>chat:write</code> scope. Then invite the bot
              to your desired channel.
            </p>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-text-field
              v-model="channel.params.token"
              label="Bot Token"
              placeholder="xoxb-your-bot-token"
              outlined
              dense
              hide-details="auto"
              :rules="rules.token"
              type="password"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-text-field
              v-model="channel.params.channel"
              label="Channel/User"
              placeholder="#general, @username, or channel ID"
              hint="Channel name (#channel), user (@user), or channel ID (C1234567890)"
              persistent-hint
              outlined
              dense
              :rules="rules.channel"
            />
          </v-col>
        </v-row>
      </template>

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
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useNotifChannelManager, SlackNotifChannel } from '@/alerting/use-notif-channels'

// Misc
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

    // Set default auth method if not specified
    if (!props.channel.params.authMethod) {
      props.channel.params.authMethod = 'webhook'
    }

    // Watch for auth method changes to clear irrelevant fields
    watch(
      () => props.channel.params.authMethod,
      (newMethod, oldMethod) => {
        if (oldMethod && oldMethod !== newMethod) {
          if (newMethod === 'webhook') {
            // Switched to webhook
            props.channel.params.token = ''
            props.channel.params.channel = ''
          } else {
            // Switched to token
            props.channel.params.webhookUrl = ''
          }
        }
      },
    )

    const rules = {
      name: [requiredRule],
      webhookUrl: [
        (v: any) => {
          if (props.channel.params.authMethod === 'webhook') {
            return v ? true : 'Webhook URL is required'
          }
          return true
        },
      ],
      token: [
        (v: any) => {
          if (props.channel.params.authMethod === 'token') {
            return v ? true : 'Bot token is required'
          }
          return true
        },
      ],
      channel: [
        (v: any) => {
          if (props.channel.params.authMethod === 'token') {
            return v ? true : 'Channel/user is required'
          }
          return true
        },
      ],
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
