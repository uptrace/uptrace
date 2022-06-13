<template>
  <div>
    <v-row no-gutters align="center" class="mb-n1">
      <v-col>
        <div class="d-flex justify-space-between filters">
          <div class="d-flex align-center filters" style="display: none">
            <v-btn
              v-model="isLabelBrowserOpen"
              x-small
              outlined
              class="my-2"
              label="Browse Labels"
              :color="isLabelBrowserOpen ? 'primary' : 'secondary'"
              @click="isLabelBrowserOpen = !isLabelBrowserOpen"
              >Browse Labels</v-btn
            >
            <div v-show="isLabelBrowserOpen">
              <LogLabelChip
                v-for="label in labels"
                :key="label.name"
                v-model="label.selected"
                :attr-key="label.name"
                label
                x-small
                class="ma-1"
              />
            </div>

            <!-- <LogLabelMenu
              v-for="label in labels.items"
              :key="label"
              :date-range="dateRange"
              :label="label"
              @click="$emit('click:filter', { key: label, op: $event.op, value: $event.value })"
            /> -->
          </div>

          <v-text-field
            class="limit-input"
            type="number"
            outlined
            dense
            label="Limit"
            hide-details="auto"
            :value="limit"
            @input="$emit('update:limit', $event)"
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
          dense
          clearable
          auto-grow
          hide-details="auto"
          spellcheck="false"
          @keyup.enter.stop.prevent
          @keydown.enter.stop.prevent="exitRawMode(true)"
          @keydown.esc.stop.prevent="exitRawMode(false)"
        ></v-textarea>
      </v-col>
    </v-row>
    <v-row v-show="isLabelBrowserOpen">
      <v-col>
        <div>
          <div class="mx-2">
            <!-- <v-row>
              <LogLabelChip
                v-for="label in labels"
                :key="label.name"
                v-model="label.selected"
                :attr-key="label.name"
                label
                x-small
                class="ma-1"
              />
            </v-row> -->

            <v-row>
              <template v-for="(label, idx) in labels">
                <div v-show="label.selected" :key="idx">
                  <LogLabelValuesCont
                    :key="label.name"
                    :date-range="dateRange"
                    :label="label.name"
                    @click="
                      $emit('click:filter', { key: label.name, op: $event.op, value: $event.value })
                    "
                  />
                </div>
              </template>
            </v-row>
          </div>
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import {
  defineComponent,
  shallowRef,
  computed,
  reactive,
  watch,
  PropType,
  ref,
} from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useLabels, Label } from '@/components/loki/logql'

// Components
// import LogLabelMenu from '@/components/loki/LogLabelMenu.vue'
import LogLabelChip from '@/components/loki/LogLabelChip.vue'
import LogLabelValuesCont from '@/components/loki/LogLabelValuesCont.vue'

export default defineComponent({
  name: 'Logql',
  components: { /*LogLabelMenu, */ LogLabelChip, LogLabelValuesCont },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    value: {
      type: String,
      required: true,
    },
    limit: {
      type: [Number, String],
      required: true,
    },
  },

  setup(props, ctx) {
    const { route } = useRouter()
    const internalQuery = shallowRef('')
    const isLabelBrowserOpen = ref(false)

    const labelValues = useLabels(() => {
      const { projectId } = route.value.params

      return {
        url: `/${projectId}/loki/api/v1/label`,
        params: {
          ...props.dateRange.lokiParams(),
        },
      }
    })

    const internalLabels = computed((): Label[] => {
      return labelValues.items.map((value: string): Label => {
        return { name: value, selected: false }
      })
    })

    const labels = computed((): Label[] => {
      return internalLabels.value.map((label) => reactive(label))
    })

    watch(
      () => props.value,
      (query) => {
        internalQuery.value = query
      },
      { immediate: true },
    )

    function exitRawMode(save: boolean) {
      if (save) {
        ctx.emit('input', internalQuery.value ?? '')
      } else {
        internalQuery.value = props.value
      }
    }

    return {
      internalQuery,
      labels,
      exitRawMode,
      isLabelBrowserOpen,
    }
  },
})
</script>

<style lang="scss" scoped>
.limit-input {
  max-width: 200px;
}
</style>
