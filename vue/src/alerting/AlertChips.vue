<template>
  <span>
    <v-chip
      small
      label
      @click.stop="$emit('click:chip', { key: AttrKey.alertType, value: alert.type })"
    >
      {{ alert.type }}
    </v-chip>

    <v-chip
      v-if="alert.attrs[AttrKey.spanSystem]"
      small
      label
      class="ml-1"
      @click.stop="
        $emit('click:chip', { key: AttrKey.spanSystem, value: alert.attrs[AttrKey.spanSystem] })
      "
    >
      {{ alert.attrs[AttrKey.spanSystem] }}
    </v-chip>

    <v-chip
      v-if="alert.attrs[AttrKey.spanKind] && alert.attrs[AttrKey.spanKind] !== 'internal'"
      small
      label
      class="ml-1"
      @click.stop="
        $emit('click:chip', { key: AttrKey.spanKind, value: alert.attrs[AttrKey.spanKind] })
      "
    >
      {{ alert.attrs[AttrKey.spanKind] }}
    </v-chip>

    <v-chip
      v-if="alert.attrs[AttrKey.spanOperation]"
      small
      label
      class="ml-1"
      @click.stop="
        $emit('click:chip', {
          key: AttrKey.spanOperation,
          value: alert.attrs[AttrKey.spanOperation],
        })
      "
    >
      {{ alert.attrs[AttrKey.spanOperation] }}
    </v-chip>
  </span>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { Alert } from '@/alerting/use-alerts'

// Misc
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'AlertChips',

  props: {
    alert: {
      type: Object as PropType<Alert>,
      required: true,
    },
  },

  setup(props) {
    return { AttrKey }
  },
})
</script>

<style lang="scss" scoped>
.v-chip ::v-deep .v-avatar {
  height: 12px !important;
  min-width: 12px !important;
  width: 12px !important;
}
</style>
