<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row>
        <v-col class="text-subtitle-1 text--primary">
          You can use webhooks to receive alerts in the JSON format via HTTP POST requests. The
          endpoint must respond with <code>2xx</code> status code or the request will be retried.
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
            label="Webhook URL"
            placeholder="https://mydomain.com/uptrace-webhook"
            hint="Publicly accessible HTTP endpoint"
            persistent-hint
            outlined
            dense
            :rules="rules.url"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-textarea
            v-model="channel.params.payload"
            label="Optional JSON payload"
            placeholder='{ "custom_key": "custom_value" }'
            hint="Custom user payload. Must be a valid JSON, e.g. strings must be quoted."
            persistent-hint
            outlined
            dense
            :rules="rules.payload"
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

      <v-row>
        <v-col>
          <v-divider />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <p>An example of the JSON message posted by Uptrace:</p>

          <PrismCode :code="alertExample" language="json" class="mb-4" />

          <p>The JSON message contains the following fields:</p>

          <v-simple-table>
            <tbody>
              <tr>
                <th>id</th>
                <td>
                  A unique message/notification id. Each alert may receive multiple notifications
                  for different states and events, for example, when alert is opened and closed.
                </td>
              </tr>
              <tr>
                <th>eventName</th>
                <td>
                  Possible values: <code>created</code>, <code>state-changed</code>,
                  <code>recurring</code> (for recurring errors).
                </td>
              </tr>
              <tr>
                <th>payload</th>
                <td>Payload provided by the user.</td>
              </tr>
              <tr>
                <th>createdAt</th>
                <td>Current time in the RFC3339 nano format.</td>
              </tr>
              <tr>
                <th>alert.id</th>
                <td>Alert id.</td>
              </tr>
              <tr>
                <th>alert.url</th>
                <td>An absolute URL to view the alert in Uptrace UI.</td>
              </tr>
              <tr>
                <th>alert.name</th>
                <td>Alert name.</td>
              </tr>
              <tr>
                <th>alert.type</th>
                <td>
                  Possible values: <code>metric</code> for metric monitors, <code>error</code> for
                  error monitors.
                </td>
              </tr>
              <tr>
                <th>alert.state</th>
                <td>Possible values: <code>open</code>, <code>closed</code>.</td>
              </tr>
              <tr>
                <th>alert.createdAt</th>
                <td>The time when the alert was created.</td>
              </tr>
            </tbody>
          </v-simple-table>
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
      payload: [
        (v: any) => {
          if (!v) {
            return true
          }
          try {
            JSON.parse(v)
          } catch (err) {
            return String(err)
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

      alertExample,
    }
  },
})

const alertExample = `{
  "id": "1676471814931265794",
  "eventName": "created",
  "payload": { "custom_key": "custom_value" },
  "createdAt": "2023-02-15T14:36:54.931265914Z",

  "alert": {
    "id": "123",
    "url": "https://app.uptrace.dev/alerting/1/alerts/123",
    "name": "Test message",
    "type": "metric",
    "state": "open",
    "createdAt": "2023-02-15T14:36:54.931265914Z"
  }
}
`.trim()
</script>

<style lang="scss" scoped></style>
