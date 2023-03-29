<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-row align="center">
      <v-col cols="auto" class="pr-4">
        <v-avatar color="blue darken-1" size="40">
          <span class="white--text text-h5">1</span>
        </v-avatar>
      </v-col>
      <v-col class="text-h5">Specify errors to monitor (optional)</v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-1 text--secondary">Notify on</v-col>
      <v-col>
        <v-checkbox
          v-model="monitor.params.notifyOnNewErrors"
          label="Notify on new errors"
          hide-details="auto"
          class="mt-0"
        />
        <v-checkbox
          v-model="monitor.params.notifyOnRecurringErrors"
          label="Notify on recurring errors"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-2 text--secondary">Conditions (joined with AND)</v-col>
      <v-col>
        <div class="mb-4">
          <AttrMatcher
            v-for="(matcher, i) in monitor.params.matchers"
            :key="i"
            :matcher="matcher"
            @click:remove="removeMatcher(i)"
          />
        </div>
        <v-btn outlined @click="addMatcher">Add condition</v-btn>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-divider />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="auto" class="pr-4">
        <v-avatar color="blue darken-1" size="40">
          <span class="white--text text-h5">3</span>
        </v-avatar>
      </v-col>
      <v-col class="text-h5">Select notification channels</v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3" class="text--secondary">Email notifications</v-col>
      <v-col class="d-flex align-center">
        <v-checkbox
          v-model="monitor.notifyEveryoneByEmail"
          label="Notify everyone by email"
          disabled
          hide-details="auto"
          class="mt-0"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-3 text--secondary">Slack and PagerDuty</v-col>
      <v-col cols="9" md="6">
        <v-select
          v-model="monitor.channelIds"
          multiple
          label="Notification channels"
          filled
          dense
          :items="channels.items"
          item-text="name"
          item-value="id"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-divider />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="3" class="mt-4 text--secondary">Monitor name</v-col>
      <v-col cols="9">
        <v-text-field
          v-model="monitor.name"
          label="Name"
          hint="Short name that describes the monitor"
          persistent-hint
          filled
          :rules="rules.name"
          hide-details="auto"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-spacer />
      <v-col cols="auto" class="pa-6">
        <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
        <v-btn type="submit" color="primary" :disabled="!isValid" :loading="monitorMan.pending">{{
          monitor.id ? 'Save' : 'Create'
        }}</v-btn>
      </v-col>
    </v-row>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useNotifChannels } from '@/alerting/use-notif-channels'
import { useMonitorManager, ErrorMonitor } from '@/alerting/use-monitors'
import { emptyAttrMatcher } from '@/use/attr-matcher'

// Components
import AttrMatcher from '@/components/AttrMatcher.vue'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'MonitorErrorForm',
  components: { AttrMatcher },

  props: {
    monitor: {
      type: Object as PropType<ErrorMonitor>,
      required: true,
    },
  },

  setup(props, ctx) {
    const channels = useNotifChannels(() => {
      return {}
    })

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = {
      name: [requiredRule],
    }

    const monitorMan = useMonitorManager()

    watch(
      () => props.monitor.params.matchers.length,
      (length) => {
        if (!length) {
          props.monitor.params.matchers = [emptyAttrMatcher()]
        }
      },
      { immediate: true },
    )

    function submit() {
      save().then(() => {
        ctx.emit('click:save')
      })
    }

    function save() {
      if (!form.value.validate()) {
        return Promise.reject()
      }

      const data = {
        ...props.monitor,
        params: {
          ...props.monitor.params,
          matchers: props.monitor.params.matchers.filter((m) => m.attr && m.value),
        },
      }

      if (props.monitor.id) {
        return monitorMan.updateErrorMonitor(data)
      }
      return monitorMan.createErrorMonitor(data)
    }

    function addMatcher() {
      props.monitor.params.matchers.push(emptyAttrMatcher())
    }

    function removeMatcher(i: number) {
      props.monitor.params.matchers.splice(i, 1)
    }

    return {
      channels,

      form,
      isValid,
      rules,
      submit,

      monitorMan,
      addMatcher,
      removeMatcher,
    }
  },
})
</script>

<style lang="scss" scoped></style>
