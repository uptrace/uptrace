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

export type LabelSelection = {
  name: string
  selected: boolean
  values: Array<object>
}

export type Label = {
  label: string
  selected: boolean
}
export type LabelValue = {
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

export function useLabelsSelected(labels: any) {
  const labelsFormatted = computed((): LabelSelection[] => {
    return (
      labels?.value?.map((label: string) => ({ name: label, selected: false, values: [] })) ?? [{}]
    )
  })
  return labelsFormatted
}

export function useValuesSelected(values: any) {
  const valuesFormatted = computed((): LabelValue[] => {
    return (
      values?.value?.map((value: string): LabelValue => ({ name: value, selected: false })) ?? [{}]
    )
  })
  return valuesFormatted
}

export function useLabelValues(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const values = computed((): string[] => {
    return data.value?.data ?? []
  })

  const valuesSelected = useValuesSelected(values)
  console.log(valuesSelected)
  // returns a proxyRef for view utilization

  return proxyRefs({ loading, items: values, selected: valuesSelected })
}

// add values into labels
