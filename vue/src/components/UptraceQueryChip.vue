<template>
  <div class="d-inline-block" style="max-width: 600px">
    <v-chip label :class="{ disabled: disabled }" @click="$emit('click:edit')">
      <v-icon v-if="deletable" left @click.stop="$emit('click:delete')">mdi-close</v-icon>
      <span v-if="info.keyword" class="mr-1 font-weight-medium">{{ info.keyword }}</span>
      <span class="d-inline-block text-truncate">{{ info.expr }}</span>
    </v-chip>
    <div v-if="error" class="ml-2 text-caption red--text text--darken-2 text-no-wrap text-truncate">
      {{ error }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

export default defineComponent({
  name: 'UptraceQueryChip',

  props: {
    query: {
      type: String,
      required: true,
    },
    error: {
      type: String,
      default: '',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    deletable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const info = computed(() => {
      return splitKeywordExpr(props.query)
    })
    return { info }
  },
})

const GROUP_BY = /^group\s+by\s+(.+)/i
const WHERE = /^where\s+(.+)/i

function splitKeywordExpr(s: string) {
  let m = s.match(GROUP_BY)
  if (m) {
    return { keyword: 'group by', expr: m[1] }
  }

  m = s.match(WHERE)
  if (m) {
    return { keyword: 'where', expr: m[1] }
  }

  return { keyword: '', expr: s }
}
</script>

<style lang="scss" scoped>
.disabled {
  opacity: 0.5;
}
</style>
