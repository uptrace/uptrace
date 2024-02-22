<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-row>
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>{{ system }} groups</v-toolbar-title>
            <v-spacer />
            <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
          </v-toolbar>

          <v-container fluid>
            <v-row>
              <v-col>
                <ApiErrorCard v-if="groups.error" :error="groups.error" />
                <PagedGroupsCard
                  v-else
                  :date-range="dateRange"
                  :systems="[system]"
                  :loading="groups.loading"
                  :groups="groups.items"
                  :columns="groups.columns"
                  :plottable-columns="groups.plottableColumns"
                  :plotted-columns="[AttrKey.spanCountPerMin]"
                  :order="groups.order"
                  :axios-params="groups.axiosParams"
                />
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute, useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor, injectQueryStore, provideQueryStore, UseUql } from '@/use/uql'
import { UseSystems } from '@/tracing/system/use-systems'
import { useGroups } from '@/tracing/use-explore-spans'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import PagedGroupsCard from '@/tracing/PagedGroupsCard.vue'

// Misc
import { AttrKey, isSpanSystem } from '@/models/otel'

export default defineComponent({
  name: 'OverviewGroups',
  components: { ApiErrorCard, PagedGroupsCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()
    const { where } = injectQueryStore()

    const system = computed(() => {
      return route.value.params.system
    })

    const query = computed(() => {
      return createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId, isSpanSystem(system.value))
        .add(where.value)
        .toString()
    })
    provideQueryStore({ query, where })

    const groups = useGroups(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: system.value,
        query: query.value,
      }
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...groups.order.queryParams(),
          system: system.value,
          query: query.value,
        },
      }
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('sort_by', `per_min(${AttrKey.spanCount})`)
        queryParams.setDefault('sort_desc', true)

        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        props.uql.parseQueryParams(queryParams)
        groups.order.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
          ...groups.order.queryParams(),
        }
      },
    })

    return {
      AttrKey,
      system,
      groups,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
