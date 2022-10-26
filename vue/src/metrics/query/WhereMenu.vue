<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text :disabled="disabled" class="v-btn--filter" v-bind="attrs" v-on="on">
        Where
      </v-btn>
    </template>

    <v-card>
      <v-list dense>
        <template v-for="metric in metrics">
          <v-menu :key="metric.id" open-on-hover offset-x transition="slide-x-transition">
            <template #activator="{ on, attrs }">
              <v-list-item v-bind="attrs" v-on="on">
                <v-list-item-content>
                  <v-list-item-title>{{ metric.name }} AS ${{ metric.alias }}</v-list-item-title>
                </v-list-item-content>
                <v-list-item-icon class="align-self-center">
                  <v-icon>mdi-menu-right</v-icon>
                </v-list-item-icon>
              </v-list-item>
            </template>

            <WhereSuggestions
              :axios-params="axiosParams"
              :metric="metric"
              :alias="metric.alias"
              @click:where="whereEqual($event)"
              @click:where-not="whereNotEqual($event)"
            />
          </v-menu>
        </template>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { AxiosParams } from '@/use/axios'
import { Suggestion } from '@/use/suggestions'
import { UseUql } from '@/use/uql'
import { Metric } from '@/metrics/types'

// Components
import WhereSuggestions from '@/metrics/query/WhereSuggestions.vue'

interface WhereSuggestion extends Suggestion {
  key: string
  value: string
}

export default defineComponent({
  name: 'WhereFilterMenu',
  components: { WhereSuggestions },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    function whereEqual(suggestion: WhereSuggestion) {
      where(suggestion, '=')
    }

    function whereNotEqual(suggestion: WhereSuggestion) {
      where(suggestion, '!=')
    }

    function where(suggestion: WhereSuggestion, op: string) {
      const editor = props.uql.createEditor()
      editor.where(suggestion.key, op, suggestion.value)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      menu,

      whereEqual,
      whereNotEqual,
    }
  },
})
</script>

<style lang="scss" scoped></style>
