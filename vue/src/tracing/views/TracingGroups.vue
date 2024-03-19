<template>
  <div>
    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="bg--primary">
            <slot name="search-filter" />

            <v-spacer />

            <div class="text-body-2">
              <span v-if="groups.hasMore">more than </span>
              <strong><NumValue :value="numGroup" verbose /></strong> groups
            </div>
          </v-toolbar>

          <v-container fluid>
            <v-row>
              <v-col>
                <ApiErrorCard v-if="groups.error" :error="groups.error" />
                <PagedGroupsCard
                  v-else
                  :date-range="dateRange"
                  :systems="systems.activeSystems"
                  :uql="uql"
                  :loading="groups.loading"
                  :groups="groups.items"
                  :columns="groups.columns"
                  :plottable-columns="groups.plottableColumns"
                  :plotted-columns="plottedColumns"
                  show-plotted-column-items
                  :order="groups.order"
                  :axios-params="groups.axiosParams"
                  @update:plotted-columns="plottedColumns = $event"
                  @update:num-group="numGroup = $event"
                />
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, watchEffect, onMounted, PropType } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { createQueryEditor, UseUql } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import PagedGroupsCard from '@/tracing/PagedGroupsCard.vue'

// Misc
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'TracingGroups',
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
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    searchInput: {
      type: String,
      default: '',
    },
  },

  setup(props, ctx) {
    props.dateRange.roundUp()

    const groups = useGroups(() => {
      return props.axiosParams
    })
    const numGroup = shallowRef(0)

    const plottedColumns = shallowRef<string[]>()
    watchEffect(() => {
      if (!groups.plottableColumns.length) {
        plottedColumns.value = undefined
        return
      }

      if (!plottedColumns.value) {
        plottedColumns.value = groups.plottableColumns.slice(0, 1).map((col) => col.name)
        return
      }

      plottedColumns.value = plottedColumns.value.filter((colName) => {
        return groups.plottableColumns.findIndex((item) => item.name === colName) >= 0
      })
    })

    onMounted(() => {
      watch(
        () => props.systems.activeSystems,
        (activeSystems) => {
          if (!activeSystems.length) {
            return
          }
          if (props.uql.query) {
            return
          }
          props.uql.query = createQueryEditor()
            .exploreAttr(AttrKey.spanGroupId, props.systems.isSpan)
            .toString()
        },
        { immediate: true },
      )
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        groups.order.parseQueryParams(queryParams)

        if (!queryParams.has('query') && props.systems.activeSystems.length) {
          queryParams.set(
            'query',
            createQueryEditor().exploreAttr(AttrKey.spanGroupId, props.systems.isSpan).toString(),
          )
        }
        props.uql.parseQueryParams(queryParams)

        if (queryParams.has('column')) {
          plottedColumns.value = queryParams.array('column')
        } else {
          plottedColumns.value = undefined // accompanied with watchEffect
        }

        const search = queryParams.string('search')
        if (search) {
          ctx.emit('update:search-input', search)
        }
      },
      toQuery() {
        const queryParams: Record<string, any> = {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
          ...groups.order.queryParams(),
        }
        if (plottedColumns.value) {
          queryParams.column = plottedColumns.value.length ? plottedColumns.value : null
        }
        if (props.searchInput) {
          queryParams.search = props.searchInput
        }
        return queryParams
      },
    })

    watch(
      () => groups.queryInfo,
      (queryInfo) => {
        if (queryInfo) {
          props.uql.setQueryInfo(queryInfo)
        }
      },
    )

    return {
      groups,

      numGroup,
      plottedColumns,
    }
  },
})
</script>

<style lang="scss" scoped></style>
