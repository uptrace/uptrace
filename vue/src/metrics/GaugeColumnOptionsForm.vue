<template>
  <div>
    <PanelSection title="Unit">
      <UnitSelect v-model="column.unit" />
    </PanelSection>

    <PanelSection title="Gauge value">
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
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Components
import PanelSection from '@/components/PanelSection.vue'
import UnitSelect from '@/metrics/UnitSelect.vue'

// Misc
import { aggFuncItems, GaugeColumn } from '@/metrics/types'
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'GaugeColumnOptionsForm',
  components: { PanelSection, UnitSelect },

  props: {
    column: {
      type: Object as PropType<GaugeColumn>,
      required: true,
    },
  },

  setup(props) {
    const rules = { aggFunc: [requiredRule] }
    return { rules, aggFuncItems }
  },
})
</script>

<style lang="scss" scoped></style>
