<template>
  <XPlaceholder>
    <v-row>
      <v-col>
        <Logql
          v-model="query"
          :date-range="dateRange"
          :limit.sync="limit"
          @click:filter="onClickFilter"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>LogQL</span>
            </v-toolbar-title>

            <v-spacer />

            <div v-if="logql.numItemInStreams" class="text-body-2 blue-grey--text text--darken-3">
              <strong><XNum :value="logql.numItemInStreams" verbose /></strong> logs
            </div>
          </v-toolbar>

          <v-row>
            <v-col>
              <LogsTable
                v-if="logql.resultType === logql.ResultType.Streams"
                :loading="logql.loading"
                :streams="logql.streams"
                @click:filter="onClickFilter"
              />
              <LogqlChart
                v-else-if="logql.resultType === logql.ResultType.Matrix"
                :date-range="dateRange"
                :loading="logql.loading"
                :result="logql.result"
                class="my-4"
              />
              <v-card v-else flat class="text-center">
                <v-card-text class="py-16">The query is empty.</v-card-text>
              </v-card>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

// Composables
import { useRouter, useQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/use/systems'
import { useLogql } from '@/components/loki/logql'

// Components
import Logql from '@/components/loki/Logql.vue'
import LogsTable from '@/components/loki/LogsTable.vue'
import LogqlChart from '@/components/loki/LogqlChart.vue'

// Utilities
import { decodeQuery } from '@/util/loki'

interface Filter {
  key: string
  op: string
  value: string
  selected: boolean
  label: string
  labels: []
  labelValues: []
}

export default defineComponent({
  name: 'LokiLogs',
  components: { Logql, LogsTable, LogqlChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const query = shallowRef('')
    const limit = shallowRef(1000)
    const logql = useLogql(() => {
      if (!query.value) {
        return undefined
      }

      const { projectId } = route.value.params
      return {
        url: `/${projectId}/loki/api/v1/query_range`,
        params: {
          ...props.dateRange.lokiParams(),
          query: query.value,
          direction: 'BACKWARD',
          limit: limit.value,
        },
      }
    })

    useQuery().sync({
      fromQuery(routerQuery) {
        if (typeof routerQuery.logql === 'string') {
          query.value = routerQuery.logql
        }
      },
      toQuery() {
        return {
          logql: query.value,
        }
      },
    })

    function onClickFilter(filter: Filter) {
      query.value = decodeQuery(query.value, filter.label, filter.value, filter.op)
    }

    return { query, limit, logql, onClickFilter }
  },
})
</script>

<style lang="scss" scoped></style>
