<template>
  <div>
    <v-menu v-model="menu" offset-y :close-on-content-click="false">
      <template #activator="{ on, attrs }">
        <v-btn text :disabled="disabled" class="v-btn--filter" v-bind="attrs" v-on="on">
          Where
        </v-btn>
      </template>

      <v-card>
        <v-list dense>
          <v-list-item
            v-for="metric in metrics"
            :key="metric.alias"
            @click="
              menu = false
              activeMetric = metric
              drawer = true
            "
          >
            <v-list-item-content>
              <v-list-item-title>${{ metric.alias }} as {{ metric.name }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-card>
    </v-menu>

    <v-navigation-drawer
      v-model="drawer"
      v-click-outside="{
        handler: onClickOutside,
        closeConditional,
      }"
      app
      temporary
      stateless
      width="500"
    >
      <FacetList
        v-if="activeMetric"
        component="metrics"
        :uql="uql"
        :axios-params="facetParams"
        :attr-prefix="`$${activeMetric.alias}.`"
        @input="drawer = $event"
      />
    </v-navigation-drawer>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { AxiosParams } from '@/use/axios'
import { UseUql } from '@/use/uql'
import { ActiveMetric as Metric } from '@/metrics/types'

// Components
import FacetList from '@/components/facet/FacetList.vue'

export default defineComponent({
  name: 'MetricsWhereMenu',
  components: { FacetList },

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
    const drawer = shallowRef(false)

    function onClickOutside() {
      drawer.value = false
    }

    function closeConditional() {
      return drawer.value
    }

    const activeMetric = shallowRef<Metric>()

    const facetParams = computed(() => {
      if (!drawer.value) {
        return null
      }
      return {
        ...props.axiosParams,
        query: props.uql.whereQuery,
        metric: activeMetric.value?.name,
      }
    })

    return {
      menu,

      drawer,
      onClickOutside,
      closeConditional,

      activeMetric,
      facetParams,
    }
  },
})
</script>

<style lang="scss" scoped></style>
