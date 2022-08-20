<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        Group by
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

            <GroupBySuggestions
              :axios-params="axiosParams"
              :metric="metric"
              :alias="metric.alias"
              @input="groupBy($event)"
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
import { Metric } from '@/metrics/use-metrics'

// Components
import GroupBySuggestions from '@/metrics/query/GroupBySuggestions.vue'

export default defineComponent({
  name: 'MetricGroupByMenu',
  components: { GroupBySuggestions },

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

    function groupBy(sugg: Suggestion) {
      const editor = props.uql.createEditor()
      editor.add(`group by ${sugg.text}`)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      menu,

      groupBy,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-select.fit {
  min-width: min-content !important;
}

.v-select.fit .v-select__selection--comma {
  text-overflow: unset;
}

.no-transform ::v-deep .v-btn {
  padding: 0 12px !important;
  text-transform: none;
}
</style>
