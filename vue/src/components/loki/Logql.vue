<template>
  <div>
    <v-row no-gutters align="center" class="mb-n1">
      <v-col>
        <div class="d-flex justify-space-between filters">
          <div class="d-flex filters" style="display: none">
            <v-btn
              x-small
              outlined
              class="my-2"
              :color="labelBrowserOpen ? 'primary' : 'secondary'"
              @click="setLabelBrowserOpen"
              v-model="labelBrowserOpen"
              label="Browse Labels"
              >Browse Labels</v-btn
            >

            <LogLabelMenu
              v-for="label in labels.items"
              :key="label"
              :date-range="dateRange"
              :label="label"
              @click="$emit('click:filter', { key: label, op: $event.op, value: $event.value })"
            />
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
    <v-row v-if="labelBrowserOpen">
      <v-col>
        <div>
          <div class="mx-2">
            <v-row>
              <LogLabelSelect
                v-for="label in labels.selected"
                :key="label.name"
                :date-range="dateRange"
                :label="label"
                x-small
                class="ma-1"
                @click:labelSelected="onLabelSelected"
              />
            </v-row>

            <v-row>
              <LogLabelValuesCont
                v-for="(label, index) in labelsSelection.labelsList"
                :key="index"
                :date-range="dateRange"
                :label="label"
                @click="
                  $emit('click:filter', { key: label.label, op: $event.op, value: $event.value })
                "
              />
            </v-row>
          </div>
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType, ref } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useLabels, Label } from '@/components/loki/logql'

// Components
import LogLabelMenu from '@/components/loki/LogLabelMenu.vue'
import LogLabelSelect from '@/components/loki/LogLabelSelect.vue'
import LogLabelValuesCont from '@/components/loki/LogLabelValuesCont.vue'

export default defineComponent({
  name: 'Logql',
  components: { LogLabelMenu, LogLabelSelect, LogLabelValuesCont },

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
    let labelsList: Label[] = []
    const labelsSelection = ref({ labelsList })
    const panel = ref([0])
    const labelBrowserOpen = ref(false)
    const labels = useLabels(() => {
      const { projectId } = route.value.params

      return {
        url: `/${projectId}/loki/api/v1/label`,
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

    function exitRawMode(save: boolean) {
      if (save) {
        ctx.emit('input', internalQuery.value ?? '')
      } else {
        internalQuery.value = props.value
      }
    }
    function setLabelBrowserOpen() {
      return (labelBrowserOpen.value = labelBrowserOpen.value ? false : true)
    }
    function onLabelSelected(value: Label) {
      if (labelsSelection.value.labelsList.some((label) => value.label === label.label)) {
        let filtered = labelsSelection.value.labelsList.filter(
          (f) => f.label !== value.label && !value.selected,
        )
        labelsSelection.value.labelsList = filtered
      } else {
        labelsSelection.value.labelsList.push(value)
      }
    }

    return {
      internalQuery,
      labels,
      exitRawMode,
      labelBrowserOpen,
      setLabelBrowserOpen,
      onLabelSelected,
      labelsSelection,
      panel,
    }
  },
})
</script>

<style lang="scss" scoped>
.limit-input {
  max-width: 200px;
}
</style>
