<template>
  <v-dialog width="auto">
    <template #activator="{ on, attrs }">
      <v-chip outlined label title="Change color" v-bind="attrs" v-on="on">
        <v-avatar :color="column.color" size="10" left></v-avatar>
        <span>{{ name }}</span>

        <UnitPicker v-model="column.unit" target-class="mr-n4" />
      </v-chip>
    </template>

    <v-card>
      <v-container fluid>
        <v-row>
          <v-col>
            <v-color-picker
              v-model="column.color"
              mode="hexa"
              show-swatches
              swatches-max-height="100%"
            ></v-color-picker>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Components
import UnitPicker from '@/components/UnitPicker.vue'

// Utilitiers
import { MetricColumn } from '@/metrics/types'

export default defineComponent({
  name: 'MetricColumnChip',
  components: { UnitPicker },

  props: {
    name: {
      type: String,
      required: true,
    },
    column: {
      type: Object as PropType<MetricColumn>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped>
.v-avatar {
  height: 18px !important;
  min-width: 18px !important;
  width: 18px !important;
}
</style>
