import { computed, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'

// Misc
import { AttrKey } from '@/models/otel'

interface ServiceGraphStats {
  durationMin: number
  durationMax: number
  durationSum: number
  durationAvg: number
  count: number
  rate: number
  errorCount: number
  errorRate: number
}

export interface ServiceGraphNode extends ServiceGraphStats {
  id: string
  name: string
  attr: string
}

export interface ServiceGraphEdge extends ServiceGraphStats {
  type: EdgeType

  clientId: string
  clientName: string

  serverId: string
  serverName: string
  serverAttr: string
}

export enum EdgeType {
  Unset = 'unset',
  Http = 'http',
  Db = 'db',
  Messaging = 'messaging',
}

export enum NodeSizeMode {
  Incoming = 'incoming',
  Outgoing = 'outgoing',
}

export enum NodeSizeMetric {
  Rate = 'rate',
  Duration = 'duration',
}

export type UseServiceGraph = ReturnType<typeof useServiceGraph>

export function useServiceGraph(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/tracing/${projectId}/service-graph`,
      params: axiosParamsSource(),
    }
  })

  const edges = computed((): ServiceGraphEdge[] => {
    const edges: ServiceGraphEdge[] = data.value?.edges ?? []
    return edges.map((edge) => {
      let serverId = edge.serverName
      if (edge.serverAttr && edge.serverAttr !== AttrKey.spanSystem) {
        serverId = edge.serverAttr + '=' + edge.serverName
      }

      return {
        ...edge,
        clientId: edge.clientName,
        serverId,
      }
    })
  })

  return proxyRefs({
    status,
    loading,
    reload,

    edges,
  })
}
