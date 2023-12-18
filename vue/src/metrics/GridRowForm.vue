<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit()">
    <v-container fluid>
      <v-row dense>
        <v-col cols="3" class="mt-4 text--secondary">Row title</v-col>
        <v-col cols="9">
          <v-text-field
            v-model="row.title"
            hint="Concise title that describes the row"
            persistent-hint
            dense
            filled
            :rules="rules.title"
            hide-details="auto"
          />
        </v-col>
      </v-row>

      <v-row dense>
        <v-col cols="3" class="mt-4 text--secondary">Optional description</v-col>
        <v-col cols="9">
          <v-text-field
            v-model="row.description"
            hint="Optional description or memo"
            persistent-hint
            dense
            filled
            hide-details="auto"
          />
        </v-col>
      </v-row>

      <v-row dense>
        <v-col cols="3"></v-col>
        <v-col>
          <v-checkbox v-model="row.expanded" label="Expand this row by default"></v-checkbox>
        </v-col>
      </v-row>

      <v-row>
        <v-spacer />
        <v-col cols="auto">
          <v-btn text @click="$emit('click:cancel')">Cancel</v-btn>
          <v-btn
            :loading="gridRowMan.pending"
            :disabled="!isValid"
            class="primary"
            @click="submit()"
            >Update</v-btn
          >
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useGridRowManager } from '@/metrics/use-dashboards'

// Misc
import { requiredRule } from '@/util/validation'
import { GridRow } from '@/metrics/types'

export default defineComponent({
  name: 'GridRowForm',

  props: {
    row: {
      type: Object as PropType<GridRow>,
      required: true,
    },
  },

  setup(props, ctx) {
    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = { title: [requiredRule] }

    const gridRowMan = useGridRowManager()
    function submit() {
      if (!form.value.validate()) {
        return
      }
      gridRowMan.save(props.row).then(() => {
        ctx.emit('save')
      })
    }

    return {
      form,
      isValid,
      rules,

      gridRowMan,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
