<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-row>
      <v-col class="text-subtitle-1">
        <p>
          These settings allow to configure personal email notifications for the project. They don't
          impact other users and projects.
        </p>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="8">
        <v-text-field
          v-model="user.current.email"
          label="Email"
          filled
          dense
          disabled
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-checkbox
          v-model="channel.notifyOnMetrics"
          label="Notify on metric alerts via email"
          hide-details="auto"
        />
        <v-checkbox
          v-model="channel.notifyOnNewErrors"
          label="Notify on new errors via email"
          hide-details="auto"
        />
        <v-checkbox
          v-model="channel.notifyOnRecurringErrors"
          label="Notify on recurring errors via email"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-btn :loading="man.pending" type="submit" color="primary">Save</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'
import { useUser } from '@/org/use-users'
import { useNotifChannelManager, EmailNotifChannel } from '@/alerting/use-notif-channels'

export default defineComponent({
  name: 'NotifChannelEmailForm',

  props: {
    channel: {
      type: Object as PropType<EmailNotifChannel>,
      required: true,
    },
  },

  setup(props) {
    const snackbar = useSnackbar()
    const user = useUser()
    const man = useNotifChannelManager()

    const form = shallowRef()
    const isValid = shallowRef(true)

    function submit() {
      if (!form.value.validate()) {
        return Promise.reject()
      }

      man.emailUpdate(props.channel).then(() => {
        snackbar.notifySuccess(`Email notification settings have been updated successfully`)
      })
    }

    return {
      user,
      man,

      form,
      isValid,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
