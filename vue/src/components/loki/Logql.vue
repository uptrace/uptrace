<template>
  <div>
    <v-row no-gutters align="center" class="mb-n1">
      <v-col>
        <div class="d-flex filters">
          <LogLabelMenu
            v-for="label in labels.items"
            :key="label"
            :date-range="dateRange"
            :label="label"
            @click="onClick(label, $event.op, $event.value)"
          />
        </div>
      </v-col>
    </v-row>

    <v-row align="center" dense>
      <v-col>
        <v-textarea
          v-model="internalQuery"
          rows="1"
          outlined
          clearable
          auto-grow
          hide-details="auto"
          @keyup.enter.stop.prevent
          @keydown.enter.stop.prevent="exitRawMode(true)"
          @keydown.esc.stop.prevent="exitRawMode(false)"
        ></v-textarea>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useLabels } from '@/components/loki/logql'

// Components
import LogLabelMenu from '@/components/loki/LogLabelMenu.vue'

export default defineComponent({
  name: 'Logql',
  components: { LogLabelMenu },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    value: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const { route } = useRouter()
    const internalQuery = shallowRef('')

    const labels = useLabels(() => {
      const { projectId } = route.value.params
      return {
        url: `/loki/api/${projectId}/v1/label`,
        params: {
          ...props.dateRange.lokiParams(),
        },
      }
    })

    watch(
      () => props.value,
      (query) => {
        internalQuery.value = query
      },
      { immediate: true },
    )

    function onClick(key: string, op: string, value: string) {
      value = JSON.stringify(value)
      ctx.emit('input', `{${key}${op}${value}}`)
    }

    function exitRawMode(save: boolean) {
      if (save) {
        ctx.emit('input', internalQuery.value ?? '')
      } else {
        internalQuery.value = props.value
      }
    }

    return { internalQuery, labels, onClick, exitRawMode }
  },
})
</script>

<style lang="scss" scoped></style>
