<template>
  <div>
    <v-row no-gutters align="center" class="mb-n1">
      <v-col>
        <div class="d-flex justify-space-between filters">
          <v-container class="px-0" fluid>
            <v-checkbox v-model="labelBrowserOpen" label="Browse Labels"> </v-checkbox>
          </v-container>
          <div class="d-flex filters" style="display: none">
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
        <!--  enclose this in a card  -->

        <!--  It should open on  -->
        <v-expansion-panels v-model="panel">
          <!-- first one for labels list -->
          <v-expansion-panel>
            <v-expansion-panel-header> Select Labels From List </v-expansion-panel-header>
            <v-expansion-panel-content>
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
                  v-for="(labelv, index) in labelsSelection.labelsList"
                  :key="index"
                  :date-range="dateRange"
                  :label="labelv"
                  @click="
                    $emit('click:filter', { key: labelv.label, op: $event.op, value: $event.value })
                  "
                />
              </v-row>
              <!-- add value selection event  -->
            </v-expansion-panel-content>
          </v-expansion-panel>
          <!-- second for values inside label (shoould be inside cards) -->
        </v-expansion-panels>
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

    function onLabelSelected(value: Label) {
      console.log(value)

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
