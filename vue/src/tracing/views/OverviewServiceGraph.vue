<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-row v-if="!serviceGraph.status.hasData()">
      <v-col>
        <v-skeleton-loader type="card" loading></v-skeleton-loader>
      </v-col>
    </v-row>

    <v-row v-else-if="!serviceGraph.edges.length" justify="center">
      <v-col cols="12" md="8" lg="6">
        <ServiceGraphHelpCard />
      </v-col>
    </v-row>

    <v-row v-else align="center">
      <v-col cols="9" style="height: 95vh">
        <ServiceGraphChart
          ref="graphRef"
          :loading="serviceGraph.loading"
          :edges="edges"
          :node-size-metric="nodeSizeMetric"
          :node-size-mode="nodeSizeMode"
          @click:node="activeItem = $event"
          @click:edge="activeItem = $event"
        />
      </v-col>
      <v-col cols="3">
        <v-row>
          <v-col class="text-center">
            <ServiceGraphHelpDialog />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-select
              v-model="activeEdgeTypes"
              :items="edgeTypeItems"
              multiple
              label="Edge type"
              solo
              dense
              hide-details="auto"
            >
              <template #item="{ item }">
                <v-list-item-action>
                  <v-checkbox :input-value="activeEdgeTypes.includes(item.value)"></v-checkbox>
                </v-list-item-action>
                <v-list-item-content>
                  <v-list-item-title>
                    <div v-html="item.text"></div>
                  </v-list-item-title>
                </v-list-item-content>
                <v-list-item-action>
                  <v-list-item-action-text>
                    <NumValue :value="item.count" />
                  </v-list-item-action-text>
                </v-list-item-action>
              </template>
            </v-select>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <div class="text-body-2 text--secondary">Size nodes by</div>

            <v-btn-toggle v-model="nodeSizeMode" tile group dense>
              <v-btn value="incoming">Incoming</v-btn>
              <v-btn value="outgoing">Outgoing</v-btn>
            </v-btn-toggle>

            <v-btn-toggle v-model="nodeSizeMetric" tile group dense>
              <v-btn value="duration">Duration</v-btn>
              <v-btn value="rate">Rate</v-btn>
            </v-btn-toggle>
          </v-col>
        </v-row>

        <v-row v-if="maxDuration !== minDuration">
          <v-col class="text--secondary">
            Edge duration: <DurationValue :value="durationRange[0]" /> -
            <DurationValue :value="durationRange[1]" />
            <v-range-slider
              v-model="durationRange"
              :min="minDuration"
              :max="maxDuration"
              hide-details
            >
            </v-range-slider>
          </v-col>
        </v-row>

        <v-row v-if="maxRate !== minRate">
          <v-col class="text--secondary">
            Edge call rate: <NumValue :value="rateRange[0]" /> - <NumValue :value="rateRange[1]" />
            <v-range-slider v-model="rateRange" :min="minRate" :max="maxRate" hide-details>
            </v-range-slider>
          </v-col>
        </v-row>

        <v-row v-if="maxErrorRate !== minErrorRate">
          <v-col class="text--secondary">
            Edge error rate: <NumValue :value="errorRateRange[0]" unit="utilization" /> -
            <NumValue :value="errorRateRange[1]" unit="utilization" />
            <v-range-slider
              v-model="errorRateRange"
              :min="minErrorRate"
              :max="maxErrorRate"
              :step="errorRateStep"
              hide-details
            >
            </v-range-slider>
          </v-col>
        </v-row>

        <template v-if="activeItem">
          <v-row>
            <v-col>
              <v-divider />
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <v-btn small @click="reset">
                <v-icon left>mdi-close-thick</v-icon>
                Reset selection
              </v-btn>
            </v-col>
          </v-row>

          <v-row class="mb-n8">
            <v-col v-if="tracingGroupsRoute" cols="auto">
              <v-btn :to="tracingGroupsRoute" depressed small>Explore groups</v-btn>
            </v-col>
            <v-col v-if="monitorMenuItems.length" cols="auto">
              <v-menu offset-y>
                <template #activator="{ on, attrs }">
                  <v-btn depressed small v-bind="attrs" v-on="on">
                    <span>Monitor</span>
                    <v-icon right>mdi-menu-down</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item
                    v-for="(item, index) in monitorMenuItems"
                    :key="index"
                    :to="item.route"
                  >
                    <v-list-item-title>{{ item.title }}</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <v-simple-table>
                <tbody>
                  <tr v-if="activeItem.name">
                    <th>Node</th>
                    <td>{{ activeItem.name }}</td>
                  </tr>
                  <template v-else>
                    <tr>
                      <th>Edge type</th>
                      <td>{{ activeItem.type }}</td>
                    </tr>
                    <tr>
                      <th>Client</th>
                      <td>{{ activeItem.clientName }}</td>
                    </tr>
                    <tr>
                      <th>Server</th>
                      <td>{{ activeItem.serverName }}</td>
                    </tr>
                  </template>
                  <tr>
                    <th>Calls per min</th>
                    <td><NumValue :value="activeItem.rate" format="verbose" /></td>
                  </tr>
                  <tr>
                    <th>Err. rate</th>
                    <td><PctValue :a="activeItem.errorCount" :b="activeItem.count" /></td>
                  </tr>
                  <tr>
                    <th>Avg duration</th>
                    <td><DurationValue :value="activeItem.durationAvg" /></td>
                  </tr>
                  <tr>
                    <th>Min duration</th>
                    <td><DurationValue :value="activeItem.durationMin" /></td>
                  </tr>
                  <tr>
                    <th>Max duration</th>
                    <td><DurationValue :value="activeItem.durationMax" /></td>
                  </tr>
                </tbody>
              </v-simple-table>
            </v-col>
          </v-row>
        </template>

        <v-row v-else>
          <v-col class="text--secondary"> Click on a node or link to view details... </v-col>
        </v-row>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { injectQueryStore, createQueryEditor, joinQuery, UseUql } from '@/use/uql'
import { UseSystems } from '@/tracing/system/use-systems'
import {
  useServiceGraph,
  ServiceGraphNode,
  ServiceGraphEdge,
  NodeSizeMode,
  NodeSizeMetric,
} from '@/tracing/use-service-graph'

// Components
import ServiceGraphChart from '@/tracing/ServiceGraphChart.vue'
import ServiceGraphHelpDialog from '@/tracing/ServiceGraphHelpDialog.vue'
import ServiceGraphHelpCard from '@/tracing/ServiceGraphHelpCard.vue'

// Misc
import { SystemName, AttrKey } from '@/models/otel'
import { defaultMetricAlias } from '@/metrics/use-metrics'
import { MINUTE } from '@/util/fmt/date'
import { quote } from '@/util/string'

export default defineComponent({
  name: 'OverviewServiceGraph',
  components: { ServiceGraphChart, ServiceGraphHelpDialog, ServiceGraphHelpCard },

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
    useTitle('Service Graph')
    const { where } = injectQueryStore()

    const nodeSizeMode = shallowRef(NodeSizeMode.Incoming)
    const nodeSizeMetric = shallowRef(NodeSizeMetric.Duration)

    const graphRef = shallowRef()
    const activeItem = shallowRef()
    function reset() {
      graphRef.value?.reset()
      activeItem.value = undefined
    }

    const serviceGraph = useServiceGraph(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.systems.axiosParams(),
        query: where.value,
      }
    })

    const activeEdgeTypes = shallowRef<string[]>([])
    const edgeTypeItems = computed(() => {
      const map = new Map<string, number>()

      for (let edge of serviceGraph.edges) {
        const count = map.get(edge.type) ?? 0
        map.set(edge.type, count + 1)
      }

      const items = []

      for (let [key, value] of map) {
        items.push({
          text: key,
          value: key,
          count: value,
        })
      }

      return items
    })
    watch(edgeTypeItems, (items) => {
      activeEdgeTypes.value = items.map((item) => item.value)
    })

    const durationRange = shallowRef([0, 0])
    const minDuration = computed(() => {
      let min = Number.MAX_VALUE
      for (let edge of serviceGraph.edges) {
        if (edge.durationAvg < min) {
          min = edge.durationAvg
        }
      }
      if (min !== Number.MAX_VALUE) {
        return Math.floor(min)
      }
      return 0
    })
    const maxDuration = computed(() => {
      let max = 0
      for (let edge of serviceGraph.edges) {
        if (edge.durationAvg > max) {
          max = edge.durationAvg
        }
      }
      return Math.ceil(max)
    })
    watch(
      () => [minDuration.value, maxDuration.value],
      (range) => {
        durationRange.value = range
      },
    )

    const rateRange = shallowRef([0, 0])
    const minRate = computed(() => {
      let min = Number.MAX_VALUE
      for (let edge of serviceGraph.edges) {
        if (edge.rate < min) {
          min = edge.rate
        }
      }
      if (min !== Number.MAX_VALUE) {
        return Math.floor(min)
      }
      return 0
    })
    const maxRate = computed(() => {
      let max = 0
      for (let edge of serviceGraph.edges) {
        if (edge.rate > max) {
          max = edge.rate
        }
      }
      return Math.ceil(max)
    })
    watch(
      () => [minRate.value, maxRate.value],
      (range) => {
        rateRange.value = range
      },
    )

    const errorRateRange = shallowRef([0, 0])
    const minErrorRate = computed(() => {
      let min = Number.MAX_VALUE
      for (let edge of serviceGraph.edges) {
        if (edge.errorRate < min) {
          min = edge.errorRate
        }
      }
      if (min !== Number.MAX_VALUE) {
        return min
      }
      return 0
    })
    const _maxErrorRate = computed(() => {
      let max = 0
      for (let edge of serviceGraph.edges) {
        if (edge.errorRate > max) {
          max = edge.errorRate
        }
      }
      return max
    })
    const errorRateStep = computed(() => {
      const delta = _maxErrorRate.value - minErrorRate.value
      if (delta >= 0.1) {
        return 0.01
      }
      return 0.001
    })
    const maxErrorRate = computed(() => {
      const prec = 1 / errorRateStep.value
      return Math.round((_maxErrorRate.value + Number.EPSILON) * prec) / prec
    })
    watch(
      () => [minErrorRate.value, maxErrorRate.value],
      (range) => {
        errorRateRange.value = range
      },
    )

    const edges = computed(() => {
      return serviceGraph.edges.filter((edge) => {
        return (
          activeEdgeTypes.value.includes(edge.type) &&
          edge.durationAvg >= durationRange.value[0] &&
          edge.durationAvg <= durationRange.value[1] &&
          edge.rate >= rateRange.value[0] &&
          edge.rate <= rateRange.value[1] &&
          edge.errorRate >= errorRateRange.value[0] &&
          edge.errorRate <= errorRateRange.value[1]
        )
      })
    })

    const tracingGroupsRoute = computed(() => {
      if (!activeItem.value) {
        return undefined
      }

      if (activeItem.value.name) {
        return tracingGroupsRouteForNode(activeItem.value as ServiceGraphNode)
      }
      return tracingGroupsRouteForLink(activeItem.value as ServiceGraphEdge)
    })

    function tracingGroupsRouteForNode(node: ServiceGraphNode) {
      const routeQuery: Record<string, any> = {}
      const query = createQueryEditor(where.value).exploreAttr(AttrKey.spanGroupId, true)

      if (node.attr === AttrKey.spanSystem) {
        routeQuery.system = node.name
      } else if (node.attr) {
        query.where(node.attr, '=', node.name)
      } else {
        routeQuery.system = SystemName.SpansAll
        query.where(AttrKey.serviceName, '=', node.name)
      }

      return {
        name: 'SpanGroupList',
        query: {
          ...routeQuery,
          query: query.toString(),
        },
      }
    }

    function tracingGroupsRouteForLink(link: ServiceGraphEdge) {
      if (!link.serverAttr) {
        return undefined
      }

      const routeQuery: Record<string, any> = {}
      const query = createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId)
        .add(where.value)
        .where(AttrKey.serviceName, '=', link.clientName)

      if (link.serverAttr === AttrKey.spanSystem) {
        routeQuery.system = link.serverName
      } else if (link.serverAttr) {
        query.where(link.serverAttr, '=', link.serverName)
      }

      return {
        name: 'SpanGroupList',
        query: {
          ...routeQuery,
          query: query.toString(),
        },
      }
    }

    const monitorMenuItems = computed(() => {
      if (!activeItem.value) {
        return []
      }

      if (activeItem.value.name) {
        const node = activeItem.value as ServiceGraphNode
        return [
          monitorMenuItemFor(
            'Monitor number of requests',
            `${node.name} number of requests`,
            'uptrace.service_graph.client_duration',
            `count($client_duration{server=${quote(node.name)}})`,
          ),
          monitorMenuItemFor(
            'Monitor number of failed requests',
            `${node.name} number of failed requests`,
            'uptrace.service_graph.failed_requests',
            `count(failed_requests{server=${quote(node.name)}})`,
          ),
          monitorMenuItemFor(
            'Monitor client-side duration',
            `${node.name} client-side duration`,
            'uptrace.service_graph.client_duration',
            `avg($client_duration{server=${quote(node.name)}})`,
          ),
          monitorMenuItemFor(
            'Monitor server-side duration',
            `${node.name} server-side duration`,
            'uptrace.service_graph.server_duration',
            `avg($server_duration{server=${quote(node.name)}})`,
          ),
        ]
      }

      const link = activeItem.value as ServiceGraphEdge
      return [
        monitorMenuItemFor(
          'Monitor number of calls',
          `${link.clientName} → ${link.serverName} number of calls`,
          'uptrace.service_graph.client_duration',
          joinQuery([
            `count($client_duration{client=${quote(link.clientName)}, server=${quote(
              link.serverName,
            )}})`,
          ]),
        ),
        monitorMenuItemFor(
          'Monitor number of failed requests',
          `${link.clientName} → ${link.serverName} number of calls`,
          'uptrace.service_graph.failed_requests',
          joinQuery([
            `count($failed_requests{client=${quote(link.clientName)}, server=${quote(
              link.serverName,
            )}})`,
          ]),
        ),
        monitorMenuItemFor(
          'Monitor client-side duration',
          `${link.clientName} → ${link.serverName} client-side duration`,
          'uptrace.service_graph.client_duration',
          joinQuery([
            `avg($client_duration{client=${quote(link.clientName)}, server=${quote(
              link.serverName,
            )}})`,
          ]),
        ),
        monitorMenuItemFor(
          'Monitor server-side duration',
          `${link.clientName} → ${link.serverName} server-side duration`,
          'uptrace.service_graph.server_duration',
          joinQuery([
            `count($server_duration{client=${quote(link.clientName)}, server=${quote(
              link.serverName,
            )}})`,
          ]),
        ),
      ]
    })

    function monitorMenuItemFor(
      title: string,
      monitorName: string,
      metricName: string,
      query: string,
    ) {
      return {
        title,
        route: {
          name: 'MonitorMetricNew',
          query: {
            name: monitorName,
            metric: metricName,
            alias: defaultMetricAlias(metricName),
            query,
            time_offset: String(-10 * MINUTE),
          },
        },
      }
    }

    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        props.uql.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
        }
      },
    })

    return {
      graphRef,
      reset,

      nodeSizeMode,
      nodeSizeMetric,
      serviceGraph,
      activeEdgeTypes,
      edgeTypeItems,
      edges,
      activeItem,

      durationRange,
      minDuration,
      maxDuration,

      rateRange,
      minRate,
      maxRate,

      errorRateRange,
      minErrorRate,
      maxErrorRate,
      errorRateStep,

      tracingGroupsRoute,
      monitorMenuItems,
    }
  },
})
</script>

<style lang="scss" scoped></style>
