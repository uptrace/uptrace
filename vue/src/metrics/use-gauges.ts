import { computed, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Misc
import { fmt } from '@/util/fmt'
import { ColumnInfo, GaugeColumn, StyledGaugeColumn } from '@/metrics/types'

export function useGaugeQuery(
  axiosParamsSource: AxiosParamsSource,
  columnMap: ComputedRef<Record<string, GaugeColumn>>,
) {
  const route = useRoute()

  const { status, loading, error, data, reload } = useWatchAxios(() => {
    const axiosParams = axiosParamsSource()
    if (!axiosParams) {
      return axiosParams
    }

    const tableAgg: Record<string, string> = {}
    for (let colName in columnMap.value) {
      const col = columnMap.value[colName]
      tableAgg[colName] = col.aggFunc
    }

    const { projectId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/gauge`,
      params: {
        ...axiosParams,
        table_agg: tableAgg,
      },
    }
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const styledColumns = computed((): StyledGaugeColumn[] => {
    const items = columns.value.map((col) => {
      return {
        ...col,
        ...columnMap.value[col.name],
      }
    })
    return items
  })

  const values = computed((): Record<string, any> => {
    return data.value?.values ?? {}
  })

  const query = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  return proxyRefs({
    status,
    error,
    loading,
    reload,

    query,
    values,
    columns,
    styledColumns,
  })
}

export function formatGauge(
  values: Record<string, number>,
  columns: StyledGaugeColumn[],
  template: string,
  noData = '-',
): string {
  if (!columns.length) {
    return noData
  }

  if (template) {
    for (let col of columns) {
      const val = values[col.name]
      if (val === undefined) {
        template = template.replaceAll(varName(col.name), '-')
        continue
      }
      template = template.replaceAll(varName(col.name), fmtVar(val, col.unit))
    }
    return template
  }

  const col = columns[0]
  const val = values[col.name]
  if (val === undefined) {
    return '-'
  }
  return fmtVar(val, col.unit)
}

function varName(colName: string): string {
  return '${' + colName + '}'
}

function fmtVar(val: any, unit: string): string {
  if (unit.startsWith('{') && unit.endsWith('}')) {
    return fmt(val)
  }
  return fmt(val, unit)
}
