import { computed, proxyRefs } from '@vue/composition-api'

// Composables
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UseLogql = ReturnType<typeof useLogql>

export interface StreamsData {
  resultType: ResultType.Streams
  result: Stream[]
}

export interface MatrixData {
  resultType: ResultType.Matrix
  result: Matrix[]
}

export enum ResultType {
  Unknown = '',
  Streams = 'streams',
  Matrix = 'matrix',
}

export interface Stream {
  stream: Record<string, string>
  values: LogValue[]
}

export type LogValue = [string, string]

export interface Matrix {
  metric: Record<string, string>
  values: MatrixValue[]
}

export type MatrixValue = [number, string]

export function useLogql(reqSource: AxiosRequestSource) {
  const { loading, data: axiosData } = useWatchAxios(reqSource)

  const data = computed((): StreamsData | MatrixData | undefined => {
    return axiosData.value?.data
  })

  const resultType = computed((): ResultType => {
    return data.value?.resultType ?? ResultType.Unknown
  })

  const result = computed((): Stream[] | Matrix[] => {
    return data.value?.result ?? []
  })

  const streams = computed((): Stream[] => {
    if (!data.value) {
      return []
    }
    if (data.value.resultType === ResultType.Streams) {
      return data.value.result
    }
    return []
  })

  const numItemInStreams = computed(() => {
    return streams.value.reduce((sum, stream) => sum + stream.values.length, 0)
  })

  return proxyRefs({
    ResultType,
    loading,
    resultType,
    result,
    streams,
    numItemInStreams,
  })
}

export interface Label {
  name: string
  selected: boolean
}

export function useLabels(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const labels = computed((): string[] => {
    return data.value?.data ?? []
  })

  return proxyRefs({ loading, items: labels })
}
