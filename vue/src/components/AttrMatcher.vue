<template>
  <v-row align="start" dense>
    <v-col cols="auto">
      <v-btn icon @click="$emit('click:remove')"><v-icon>mdi-close</v-icon></v-btn>
    </v-col>
    <v-col cols="4">
      <v-text-field
        v-model="matcher.attr"
        label="Attribute name"
        placeholder="deployment.environment"
        outlined
        dense
        :rules="rules.attr"
        hide-details="auto"
      />
    </v-col>
    <v-col cols="2">
      <v-select v-model="matcher.op" :items="opItems" outlined dense hide-details="auto"></v-select>
    </v-col>
    <v-col cols="4">
      <v-text-field
        v-model="matcher.value"
        label="Attribute value"
        outlined
        dense
        :rules="rules.value"
        hide-details="auto"
      />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { AttrMatcher, AttrMatcherOp } from '@/use/attr-matcher'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'SamplingRuleMatcherRow',

  props: {
    matcher: {
      type: Object as PropType<AttrMatcher>,
      required: true,
    },
    required: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const opItems = [
      { text: '==', value: AttrMatcherOp.Equal },
      { text: '!=', value: AttrMatcherOp.NotEqual },
    ]
    const rules = computed(() => {
      const rules: Record<string, any> = {
        attr: [],
        value: [],
      }
      if (props.required) {
        rules.attr.push(requiredRule)
        rules.value.push(requiredRule)
      }
      return rules
    })

    return { opItems, rules }
  },
})
</script>

<style lang="scss" scoped></style>
