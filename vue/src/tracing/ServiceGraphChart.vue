<template>
  <EChart
    :loading="loading"
    :height="chartProps.height"
    :option="chartProps.option"
    @input="onInit"
  />
</template>

<script lang="ts">
import { escape } from 'lodash-es'
import * as echarts from 'echarts'
import colors from 'vuetify/lib/util/colors'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Components
import EChart from '@/components/EChart.vue'

// Composables
import {
  ServiceGraphNode,
  ServiceGraphEdge,
  NodeSizeMode,
  NodeSizeMetric,
} from '@/tracing/use-service-graph'

// Utilities
import { num, numShort, duration, durationShort, utilization } from '@/util/fmt'
import { mapNumber } from '@/util/mapping'

export default defineComponent({
  name: 'ServiceGraphChart',
  components: { EChart },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    edges: {
      type: Array as PropType<ServiceGraphEdge[]>,
      required: true,
    },
    nodeSizeMode: {
      type: String as PropType<NodeSizeMode>,
      required: true,
    },
    nodeSizeMetric: {
      type: String as PropType<NodeSizeMetric>,
      required: true,
    },
  },

  setup(props, ctx) {
    const activeNode = shallowRef('')
    const activeEdge = shallowRef('')
    const highlightedId = shallowRef('')
    function reset() {
      activeNode.value = ''
      activeEdge.value = ''
      highlightedId.value = ''
    }

    const edges = computed(() => {
      if (!activeNode.value) {
        return props.edges
      }

      return props.edges.filter((edge) => {
        return edge.clientId === activeNode.value || edge.serverId === activeNode.value
      })
    })

    const graphEdges = computed(() => {
      const graphEdges = []

      for (let edge of edges.value) {
        const graphEdge: Record<string, any> = {
          id: `${edge.clientId} -> ${edge.serverId}`,
          name: '',
          source: edge.clientId,
          target: edge.serverId,
          label: { show: false },
          lineStyle: { width: 0.5, opacity: 1, color: colors.green.darken2 },
          _edge: edge,
        }
        graphEdges.push(graphEdge)

        switch (props.nodeSizeMetric) {
          case NodeSizeMetric.Rate:
            graphEdge.name = numShort(edge.rate)
            break
          case NodeSizeMetric.Duration:
            graphEdge.name = durationShort(edge.durationAvg)
            break
          default:
            console.error(`unsupported node size metric: ${props.nodeSizeMetric}`)
            break
        }

        if (edge.errorRate >= 0.05) {
          graphEdge.lineStyle.color = colors.red.darken2
        } else if (edge.errorRate >= 0.005) {
          graphEdge.lineStyle.color = colors.pink.darken2
        }

        if (edges.value.length < 8) {
          graphEdge.label.show = true
          graphEdge.symbol = ['none', 'arrow']
          graphEdge.symbolSize = [0, 10]
        } else if (edges.value.length < 16) {
          graphEdge.label.show = true
        }

        const id = highlightedId.value || activeEdge.value
        if (id) {
          if (graphEdge.id === id || graphEdge.source === id || graphEdge.target === id) {
            graphEdge.label.show = true
            graphEdge.lineStyle.width = 2
            graphEdge.symbol = ['none', 'arrow']
            graphEdge.symbolSize = [0, 10]
          } else {
            graphEdge.lineStyle.opacity = 0.1
          }
        }
      }

      return graphEdges
    })

    const nodes = computed(() => {
      const nodeMap = new Map<string, ServiceGraphNode>()

      for (let edge of edges.value) {
        let client = nodeMap.get(edge.clientId)
        if (!client) {
          client = emptyNode()
          client.id = edge.clientId
          client.name = edge.clientName
          nodeMap.set(edge.clientId, client)
        }

        let server = nodeMap.get(edge.serverId)
        if (!server) {
          server = emptyNode()
          server.id = edge.serverId
          server.name = edge.serverName
          server.attr = edge.serverAttr
          nodeMap.set(edge.serverId, server)
        }

        let node: ServiceGraphNode | undefined

        switch (props.nodeSizeMode) {
          case NodeSizeMode.Incoming:
            node = server
            break
          case NodeSizeMode.Outgoing:
            node = client
            break
          default:
            console.error(`unsupported node size mode: ${props.nodeSizeMode}`)
            break
        }

        if (!node) {
          continue
        }

        if (node.count === 0) {
          node.durationMin = edge.durationMin
          node.durationMax = edge.durationMax
        } else {
          node.durationMin = Math.min(node.durationMin, edge.durationMin)
          node.durationMax = Math.min(node.durationMin, edge.durationMax)
        }
        node.durationSum += edge.durationSum
        node.count += edge.count
        node.rate += edge.rate
        node.errorCount += edge.errorCount
      }

      const nodes: ServiceGraphNode[] = []

      for (let node of nodeMap.values()) {
        nodes.push(node)

        if (node.count > 0) {
          node.errorRate = node.errorCount / node.count
          node.durationAvg = node.durationSum / node.count
        }
      }

      return nodes
    })

    const graphNodes = computed(() => {
      let maxRate = 0
      let maxDuration = 0

      for (let node of nodes.value) {
        if (node.rate > maxRate) {
          maxRate = node.rate
        }
        if (node.durationAvg > maxDuration) {
          maxDuration = node.durationAvg
        }
      }

      let graphNodes = []

      for (let node of nodes.value) {
        const graphNode = {
          id: node.id,
          name: node.id,
          symbolSize: 10,
          itemStyle: { color: colors.green.lighten2 },
          _node: node,
        }
        graphNodes.push(graphNode)

        if (node.errorRate >= 0.05) {
          graphNode.itemStyle.color = colors.red.lighten2
        } else if (node.errorRate >= 0.005) {
          graphNode.itemStyle.color = colors.pink.lighten2
        }

        switch (props.nodeSizeMetric) {
          case NodeSizeMetric.Rate:
            graphNode.symbolSize = mapNumber(node.rate, 0, maxRate, 10, 40)
            break
          case NodeSizeMetric.Duration:
            graphNode.symbolSize = mapNumber(node.durationAvg, 0, maxDuration, 10, 40)
            break
          default:
            console.error(`unsupported node size metric: ${props.nodeSizeMetric}`)
            break
        }
      }

      return graphNodes
    })

    const chartProps = computed(() => {
      const series: Record<string, any> = {
        name: 'Service Graph',
        type: 'graph',

        data: graphNodes.value,
        edges: graphEdges.value,

        roam: true, // enable zoom
        scaleLimit: {
          min: 0.4,
          max: 3,
        },

        label: {
          show: true,
          position: 'right',
          fontSize: 12,
          formatter: (params: any) => {
            return params.data._node.name
          },
        },
        edgeLabel: {
          //position: 'start',
          fontSize: 12,
          color: '#000',
          formatter: (params: any) => {
            return params.data.name
          },
        },

        lineStyle: {
          width: 0.5,
          color: 'source',
          curveness: 0.3,
        },
      }

      const chart = {
        height: '100%',
        option: {
          textStyle: {
            fontFamily: '"Roboto", sans-serif',
          },
          tooltip: {
            formatter: (params: any) => {
              let title = ''
              const rows = []

              switch (params.dataType) {
                case 'node': {
                  const node = params.data._node
                  title = `${escape(node.id)} (${props.nodeSizeMode} calls)`
                  rows.push('<tr>', '<td>Calls per min</td>', `<td>${num(node.rate)}</td>`, '</tr>')
                  rows.push(
                    '<tr>',
                    '<td>Err. rate</td>',
                    `<td>${utilization(node.errorRate)}</td>`,
                    '</tr>',
                  )
                  rows.push(
                    '<tr>',
                    '<td>Duration</td>',
                    `<td>${duration(node.durationAvg)}</td>`,
                    '</tr>',
                  )
                  break
                }
                case 'edge': {
                  const edge = params.data._edge
                  title = `${escape(edge.clientId)} &rarr; ${escape(edge.serverId)}`
                  rows.push('<tr>', '<td>Calls per min</td>', `<td>${num(edge.rate)}</td>`, '</tr>')
                  rows.push(
                    '<tr>',
                    '<td>Err. rate</td>',
                    `<td>${utilization(edge.errorRate)}</td>`,
                    '</tr>',
                  )
                  rows.push(
                    '<tr>',
                    '<td>Duration</td>',
                    `<td>${duration(edge.durationAvg)}</td>`,
                    '</tr>',
                  )
                  break
                }
                default:
                  return undefined
              }

              const ss = [
                '<div class="chart-tooltip">',
                `<p>${title}</p>`,
                '<table>',
                '<tbody>',
                rows.join(''),
                '</tbody>',
                '</table>',
                '</div',
              ]

              return ss.join('')
            },
          },
          animationDurationUpdate: 1500,
          animationEasingUpdate: 'quinticInOut',
          series: [series],
        },
      }

      // TODO: pick a better layout
      series.layout = 'circular'
      series.circular = { rotateLabel: true }

      return chart
    })

    function onInit(echart: echarts.ECharts) {
      echart.on('click', (params: any) => {
        if (params.dataType === 'node') {
          activeNode.value = params.data.id
          activeEdge.value = ''
          ctx.emit('click:node', params.data._node)
        } else {
          activeNode.value = ''
          activeEdge.value = params.data.id
          ctx.emit('click:edge', params.data._edge)
        }
        highlightedId.value = ''
      })
      echart.on('mouseover', (params: any) => {
        highlightedId.value = params.data.id
      })
      echart.on('mouseout', (params: any) => {
        highlightedId.value = ''
      })
    }

    return { chartProps, onInit, reset }
  },
})

function emptyNode(): ServiceGraphNode {
  return {
    id: '',
    name: '',
    attr: '',

    durationMin: 0,
    durationMax: 0,
    durationSum: 0,
    durationAvg: 0,
    count: 0,
    rate: 0,
    errorCount: 0,
    errorRate: 0,
  }
}
</script>

<style lang="scss" scoped></style>
