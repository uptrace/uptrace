<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on"> Duration </v-btn>
    </template>

    <v-form @submit.prevent="addFilter">
      <v-card width="360px">
        <v-card-text class="py-6">
          <v-row align="center" class="mb-n4">
            <v-col>
              <v-text-field
                v-model="gte"
                type="number"
                min="0"
                label=">= duration"
                suffix="ms"
                outlined
                dense
                autofocus
                hide-details="auto"
              />
            </v-col>
            <v-col cols="auto" class="px-0">AND</v-col>
            <v-col>
              <v-text-field
                v-model="lt"
                type="number"
                :min="gte + 1"
                label="< duration"
                suffix="ms"
                outlined
                dense
                hide-details="auto"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn type="submit" color="primary" :disabled="!isValid">Filter</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'DurationFilterMenu',

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const gte = shallowRef()
    const lt = shallowRef()

    const isValid = computed((): boolean => {
      return gte.value || lt.value
    })

    function addFilter() {
      if (!isValid.value) {
        menu.value = false
        return
      }

      const editor = props.uql.createEditor()

      if (gte.value) {
        editor.replaceOrPush(
          /^where\s+span\.duration\s+>=\s.+$/i,
          `where ${AttrKey.spanDuration} >= ${gte.value}ms`,
        )
      }
      if (lt.value) {
        editor.replaceOrPush(
          /^where\s+span\.duration\s+<\s.+$/i,
          `where ${AttrKey.spanDuration} < ${lt.value}ms`,
        )
      }

      props.uql.commitEdits(editor)

      menu.value = false
      gte.value = undefined
      lt.value = undefined
    }

    return { menu, gte, lt, isValid, addFilter }
  },
})
</script>

<style lang="scss" scoped></style>
