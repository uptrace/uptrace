<template>
  <v-btn-toggle color="blue accent-3" tile group class="mb-2" @change="$emit('input', $event)">
    <v-btn v-for="item in states" :key="item.state" :value="item.state" small>
      <MonitorStateAvatar :state="item.state" />
      <span class="ml-1">({{ item.count }})</span>
    </v-btn>
  </v-btn-toggle>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { StateCount } from '@/alerting/use-monitors'

// Components
import MonitorStateAvatar from '@/alerting/MonitorStateAvatar.vue'

export default defineComponent({
  name: 'MonitorStateCounts',
  components: { MonitorStateAvatar },

  props: {
    states: {
      type: Array as PropType<StateCount[]>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped></style>
