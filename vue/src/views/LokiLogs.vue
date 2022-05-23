<template>
  <XPlaceholder>
    <v-row>
      <v-col>
        <Logql :date-range="dateRange" :query="query" @update:query="query = $event" />
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

            <div class="text-body-2 blue-grey--text text--darken-3">
              <strong><XNum :value="logql.pager.numItem" verbose /></strong> spans
            </div>
          </v-toolbar>

          <v-row class="px-4 pb-4">
            <v-col>
              <LogsTable :loading="logql.loading" :labels="logql.labels" :logs="logql.logs" />
            </v-col>
          </v-row>
        </v-card>

        <XPagination :pager="logql.pager" />
      </v-col>
    </v-row>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

// Composables
//import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/use/systems'
import { useLogql } from '@/components/loki/logql'

// Components
import Logql from '@/components/loki/Logql.vue'
import LogsTable from '@/components/loki/LogsTable.vue'

export default defineComponent({
  name: 'LokiLogs',
  components: { Logql, LogsTable },

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
    //const { route } = useRouter()
    const query = shallowRef('{foo="bar"}')

    const logql = useLogql(() => {
      //const { projectId } = route.value.params
      return {
        url: `/loki/api/v1/query_range`,
        params: {
          ...props.dateRange.lokiParams(),
          query: query.value,
          direction: 'BACKWARD',
          limit: 1000,
        },
      }
    })

    return { query, logql }
  },
})
</script>

<style lang="scss" scoped></style>
