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

export interface Label {
  name: string
  selected: boolean
}
export interface LabelValue {
  name: string
  selected: boolean
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

export function useLabels(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const labels = computed((): string[] => {
    return data.value?.data ?? []
  })

  const labelsSelected = useLabelsSelected(labels)

  return proxyRefs({ loading, items: labels, selected: labelsSelected })
}

export function useLabelValues(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const values = computed((): string[] => {
    return data.value?.data ?? []
  })

  const valuesSelected = useValuesSelected(values)

  return proxyRefs({ loading, items: values, selected: valuesSelected })
}

export function useLabelsSelected(labels: any) {
  return computed(
    (): Label[] =>
      labels?.value?.map((label: string) => ({ name: label, selected: false })) ?? [{}],
  )
}

export function useValuesSelected(values: any) {
  return computed(
    (): LabelValue[] =>
      values?.value?.map((value: string): LabelValue => ({ name: value, selected: false })) ?? [{}],
  )
}
