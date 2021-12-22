<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>Search</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-form @submit.prevent="addFilter">
      <v-card width="400">
        <v-card-text class="pa-6">
          <v-row>
            <v-col>
              <v-text-field
                v-model="attrValue"
                label="Contains substr1|substr2|substr3"
                hint='Case-insensitive options separated with "|"'
                persistent-hint
                filled
                dense
                autofocus
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn type="submit" class="primary" :disabled="!attrValue.length">Filter</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

// Composables
import { UseUql } from '@/use/uql'

// Utilities
import { xkey } from '@/models/otelattr'
import { quote } from '@/util/string'

export default defineComponent({
  name: 'SearchFilterMenu',

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const attrValue = shallowRef('')

    function addFilter() {
      const quotedValue = quote(attrValue.value)

      const editor = props.uql.createEditor()
      editor.add(`where ${xkey.spanName} contains ${quotedValue}`)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      xkey,
      menu,
      attrValue,

      addFilter,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-btn-toggle ::v-deep .v-btn {
  text-transform: none;
}
</style>
