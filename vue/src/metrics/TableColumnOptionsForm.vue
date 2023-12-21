<template>
  <div>
    <div class="mb-4">
      <v-dialog v-model="dialog" width="auto">
        <template #activator="{ on, attrs }">
          <v-btn outlined small v-bind="attrs" v-on="on">
            <v-icon size="24" :color="column.color" left>mdi-circle</v-icon>
            <span>{{ column.color }}</span>
          </v-btn>
        </template>
        <v-card>
          <v-container fluid>
            <v-row>
              <v-col>
                <v-color-picker
                  v-model="column.color"
                  mode="hexa"
                  show-swatches
                  swatches-max-height="300"
                ></v-color-picker>
              </v-col>
            </v-row>

            <v-row>
              <v-spacer />
              <v-col cols="auto">
                <v-btn color="primary" @click="dialog = false">OK</v-btn>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-dialog>
    </div>

    <PanelSection title="Unit">
      <UnitSelect v-model="column.unit" />
    </PanelSection>

    <PanelSection title="Table value">
      <v-select
        v-model="column.aggFunc"
        :items="aggFuncItems"
        dense
        filled
        required
        :rules="rules.aggFunc"
        hide-details="auto"
      />
    </PanelSection>

    <v-checkbox v-model="column.sparklineDisabled" label="Omit sparkline" />
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Components
import PanelSection from '@/components/PanelSection.vue'
import UnitSelect from '@/metrics/UnitSelect.vue'

// Misc
import { aggFuncItems, TableColumn } from '@/metrics/types'
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'TableColumnOptionsForm',
  components: { PanelSection, UnitSelect },

  props: {
    column: {
      type: Object as PropType<TableColumn>,
      required: true,
    },
  },

  setup(props) {
    const dialog = shallowRef(false)
    const rules = { aggFunc: [requiredRule] }
    return { dialog, rules, aggFuncItems }
  },
})
</script>

<style lang="scss" scoped></style>
