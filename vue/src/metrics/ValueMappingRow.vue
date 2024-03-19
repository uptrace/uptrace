<template>
  <v-row>
    <v-col cols="auto">
      <v-btn icon @click="$emit('click:remove')">
        <v-icon>mdi-delete</v-icon>
      </v-btn>
    </v-col>
    <v-col class="d-flex align-center">
      <BtnSelectMenu v-model="value.op" :items="mappingOpItems" outlined target-class="mr-2" />
      <v-text-field
        v-model.number="value.value"
        type="number"
        :disabled="value.op === MappingOp.Any"
        label="Number"
        single-line
        outlined
        dense
        hide-details="auto"
        style="width: 50px"
      />
    </v-col>
    <v-col cols="auto" class="text-h5">&rarr;</v-col>
    <v-col>
      <v-text-field
        v-model="value.text"
        label="Text"
        single-line
        outlined
        dense
        hide-details="auto"
      />
    </v-col>
    <v-col cols="auto">
      <v-dialog v-model="dialog" width="auto">
        <template #activator="{ on, attrs }">
          <v-btn icon title="Click to change color" v-bind="attrs" v-on="on">
            <v-icon size="24" :color="value.color">mdi-circle</v-icon>
          </v-btn>
        </template>

        <v-card>
          <v-container fluid>
            <v-row>
              <v-col>
                <v-color-picker
                  v-model="value.color"
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
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Components
import BtnSelectMenu from '@/components/BtnSelectMenu.vue'

// Misc
import { ValueMapping, MappingOp } from '@/metrics/types'

export default defineComponent({
  name: 'ValueMappingRow',
  components: { BtnSelectMenu },

  props: {
    value: {
      type: Object as PropType<ValueMapping>,
      required: true,
    },
  },

  setup() {
    const dialog = shallowRef(false)

    const mappingOpItems = [
      { text: 'any', value: MappingOp.Any },
      { text: '==', value: MappingOp.Equal },
      { text: '>', value: MappingOp.Gt },
      { text: '>=', value: MappingOp.Gte },
      { text: '<', value: MappingOp.Lt },
      { text: '<=', value: MappingOp.Lte },
    ]

    return { MappingOp, dialog, mappingOpItems }
  },
})
</script>

<style lang="scss" scoped></style>
