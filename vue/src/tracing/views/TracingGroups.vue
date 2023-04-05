<template>
  <div>
    <template v-if="groups.errorCode === 'invalid_query'">
      <v-row>
        <v-col>
          <v-banner>
            <v-icon slot="icon" color="error" size="36">mdi-alert-circle</v-icon>
            <span class="subtitle-1 text--secondary">{{ groups.errorMessage }}</span>
          </v-banner>

          <PrismCode v-if="groups.query" :code="groups.query" language="sql" />
        </v-col>
      </v-row>
    </template>

    <v-row v-else>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>Groups</span>
            </v-toolbar-title>

            <v-text-field
              v-model="searchInput"
              label="Quick search over group names"
              clearable
              outlined
              dense
              hide-details="auto"
              class="ml-8"
              style="max-width: 300px"
            />

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <span v-if="groups.hasMore">more than </span>
              <strong><XNum :value="numGroup" verbose /></strong> groups
            </div>
          </v-toolbar>

          <v-container fluid>
            <GroupsList
              :date-range="dateRange"
              :events-mode="eventsMode"
              :uql="uql"
              :loading="groups.loading"
              :is-resolved="groups.status.isResolved()"
              :groups="filteredGroups"
              :columns="groups.columns"
              :plottable-columns="groups.plottableColumns"
              show-plotted-column-items
              :order="groups.order"
              :show-system="showSystem"
              :axios-params="internalAxiosParams"
              @update:num-group="numGroup = $event"
            />
          </v-container>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useGroups, Group } from '@/tracing/use-explore-spans'

// Components
import GroupsList from '@/tracing/GroupsList.vue'

import { isDummySystem } from '@/models/otel'

export default defineComponent({
  name: 'TracingGroups',
  components: { GroupsList },

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
    eventsMode: {
      type: Boolean,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const groups = useGroups(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/groups`,
        params: props.axiosParams,
      }
    })

    const searchInput = shallowRef('')
    const numGroup = shallowRef(0)
    const filteredGroups = computed((): Group[] => {
      if (!searchInput.value) {
        return groups.items
      }
      return fuzzyFilter(groups.items, searchInput.value, { key: '_name' })
    })

    const internalAxiosParams = computed(() => {
      if (!groups.status.isResolved()) {
        // Block requests until items are ready.
        return { _: undefined }
      }
      return props.axiosParams
    })

    const showSystem = computed(() => {
      if (route.value.params.eventSystem) {
        return false
      }

      const systems = props.systems.activeSystem
      if (systems.length > 1) {
        return true
      }
      if (systems.length === 1) {
        return isDummySystem(systems[0])
      }
      return false
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
      internalAxiosParams,
      groups,

      searchInput,
      numGroup,
      filteredGroups,
      showSystem,
    }
  },
})
</script>

<style lang="scss" scoped></style>
