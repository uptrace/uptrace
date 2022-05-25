<template>
  <XPlaceholder>
    <v-row>
      <v-col>
        <Logql
          :date-range="dateRange"
          v-model="query"
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
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/use/systems'
import { useLogql } from '@/components/loki/logql'

// Components
import Logql from '@/components/loki/Logql.vue'
import LogsTable from '@/components/loki/LogsTable.vue'
import LogqlChart from '@/components/loki/LogqlChart.vue'

// Utilities
import { escapeRe } from '@/util/string'

interface Filter {
  key: string
  op: string
  value: string
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
    const query = shallowRef('{foo="bar"}')
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

    function onClickFilter(filter: Filter) {
      query.value = updateQuery(query.value, filter.key, filter.op, JSON.stringify(filter.value))
    }

    return { query, limit, logql, onClickFilter }
  },
})

const STREAM_SEL_RE = /{[^}]*}/

function updateQuery(query: string, key: string, op: string, value: string): string {
  const selector = `${key}${op}${value}`

  if (!query) {
    return `{${selector}}`
  }

  const m = query.match(STREAM_SEL_RE)
  if (!m) {
    return `{${selector}}`
  }

  let found = m[0]

  const e = escapeRe
  const re = new RegExp(`${e(key)}\\s*${e(op)}\\s*("(?:[^"\\\\]|\\\\.)*")`)
  if (re.test(found)) {
    found = found.replace(re, selector)
  } else {
    found = found.slice(1, -1)
    found = `{${found}, ${selector}}`
  }

  return query.replace(STREAM_SEL_RE, found)
}
</script>

<style lang="scss" scoped></style>
