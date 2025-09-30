<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col class="text-subtitle-1 text--primary">
          <p>
            Slack notifications can be configured using either webhooks or bot tokens.
            Choose the method that best fits your setup:
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
            v-model="authMethodIndex"
            mandatory
            color="primary"
            group
            class="v-btn-group--horizontal"
          >
            <v-btn value="0" class="text-none">
              <v-icon left>mdi-webhook</v-icon>
              Webhook
            </v-btn>
            <v-btn value="1" class="text-none">
              <v-icon left>mdi-robot</v-icon>
              Bot Token
            </v-btn>
          </v-btn-toggle>
        </v-col>
      </v-row>

      <!-- Webhook Configuration -->
      <template v-if="currentAuthMethod === 'webhook'">
        <v-row>
          <v-col class="text-subtitle-2 text--primary">
            <p>
              Create an
              <a href="https://api.slack.com/messaging/webhooks" target="_blank">incoming webhook</a>
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
      <template v-if="currentAuthMethod === 'token'">
        <v-row>
          <v-col class="text-subtitle-2 text--primary">
            <p>
              Create a
              <a href="https://api.slack.com/apps" target="_blank">Slack app</a>
              and generate a bot token with the <code>chat:write</code> scope.
              Then invite the bot to your desired channel.
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
import { defineComponent, shallowRef, computed, ref, watch, PropType } from 'vue'

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
    
    // Use a reactive ref to track the current auth method
    const currentAuthMethod = ref(props.channel.params.authMethod || 'webhook')
    
    // Watch for external changes to the channel params and sync with our ref
    watch(() => props.channel.params.authMethod, (newMethod) => {
      if (newMethod && newMethod !== currentAuthMethod.value) {
        currentAuthMethod.value = newMethod
      }
    })
    
    // Computed property for button toggle index
    const authMethodIndex = computed({
      get: () => {
        return currentAuthMethod.value === 'token' ? "1" : "0"
      },
      set: (value: string) => {
        const method = value === "1" ? 'token' : 'webhook'
        const oldMethod = currentAuthMethod.value
        
        currentAuthMethod.value = method
        props.channel.params.authMethod = method
        
        // Clear fields when switching methods
        if (oldMethod !== method) {
          if (method === 'webhook') {
            // Switched to webhook
            props.channel.params.token = ''
            props.channel.params.channel = ''
          } else {
            // Switched to token
            props.channel.params.webhookUrl = ''
          }
        }
      }
    })

    const rules = {
      name: [requiredRule],
      webhookUrl: [
        (v: any) => {
          if (currentAuthMethod.value === 'webhook') {
            return v ? true : 'Webhook URL is required'
          }
          return true
        },
      ],
      token: [
        (v: any) => {
          if (currentAuthMethod.value === 'token') {
            return v ? true : 'Bot token is required'
          }
          return true
        },
      ],
      channel: [
        (v: any) => {
          if (currentAuthMethod.value === 'token') {
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
      authMethodIndex,
      currentAuthMethod,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
