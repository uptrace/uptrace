<template>
  <v-text-field
    :value="value"
    placeholder="Search: option1|option2 -exclude"
    prepend-inner-icon="mdi-magnify"
    clearable
    outlined
    dense
    hide-details="auto"
    style="max-width: 350px"
    @input="$emit('input', $event)"
  >
    <template #append-outer>
      <v-tooltip bottom max-width="500">
        <template #activator="{ on }">
          <v-icon v-on="on"> mdi-help-circle-outline </v-icon>
        </template>

        <ul class="pa-2">
          <li>
            <strong>option1|option2</strong>. Include items that contain "option1" OR "option2".
          </li>
          <li><strong>word1 word2</strong>. Include items that contain "word1" AND "word2".</li>
          <li><strong>"whole phrase"</strong>. Include items that contain "whole phrase".</li>
          <li><strong>-word</strong>. Exclude items that contain "word".</li>
          <li><strong>~regexp</strong>. Include items that match the regexp "word".</li>
        </ul>

        <p>
          You can create complex queries combining the options above, for example,
          <strong>option1|"whole phrase" -foo -bar -foo|bar ~regexp</strong>.
        </p>
      </v-tooltip>
    </template>
  </v-text-field>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'QuickSearch',

  props: {
    value: {
      type: String,
      default: '',
    },
  },
})
</script>

<style lang="scss" scoped>
li {
  margin: 4px 0;
}
</style>
