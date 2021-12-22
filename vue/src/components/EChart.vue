<template>
  <div ref="container">
    <v-sheet :width="width" :height="height" class="mx-auto echart"></v-sheet>
  </div>
</template>

<script lang="ts">
import { debounce } from 'lodash'
import * as echarts from 'echarts'
import {
  defineComponent,
  shallowRef,
  watch,
  onMounted,
  onUnmounted,
  PropType,
} from '@vue/composition-api'

type GroupName = string | symbol

export interface EChartProps {
  width?: number | string
  height: number | string
  option: echarts.EChartsOption
}

export default defineComponent({
  name: 'EChart',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Object as PropType<echarts.ECharts>,
      default: undefined,
    },
    width: {
      type: [Number, String],
      default: '100%',
    },
    height: {
      type: [Number, String],
      required: true,
    },
    option: {
      type: Object as PropType<echarts.EChartsOption>,
      default: undefined,
    },
    group: {
      type: [String, Symbol] as PropType<GroupName>,
      default: undefined,
    },
  },

  setup(props, ctx) {
    let echart: echarts.ECharts
    const container = shallowRef<HTMLElement>()

    function init() {
      if (echart) {
        return
      }

      const div = container.value!.getElementsByClassName('echart')[0] as HTMLDivElement
      echart = echarts.init(div)

      ctx.emit('input', echart)

      if (props.group) {
        register(props.group, echart)
      }
    }

    const setOption = debounce((option: echarts.EChartsOption) => {
      if (echart.isDisposed()) {
        return
      }

      echart.setOption(option, { notMerge: true, lazyUpdate: true, silent: true })
    }, 50)

    onMounted(() => {
      watch(
        () => props.option,
        (option) => {
          if (option) {
            init()
            setOption(option)
          }
        },
        { immediate: true },
      )

      watch(
        () => props.loading,
        (loading) => {
          init()
          if (loading) {
            echart.showLoading()
          } else {
            echart.hideLoading()
          }
        },
        { immediate: true },
      )
    })

    onUnmounted(() => {
      if (echart) {
        if (props.group) {
          unregister(props.group, echart)
        }
        echart.dispose()
      }
    })

    return { container }
  },
})

//------------------------------------------------------------------------------

const groupMap: Record<GroupName, echarts.ECharts[]> = {}

function register(groupName: GroupName, echart: echarts.ECharts): void {
  let group = groupMap[groupName as string]
  if (!group) {
    group = []
    groupMap[groupName as string] = group
  }
  group.push(echart)
  connect(echart, group)
}

function unregister(groupName: GroupName, echart: echarts.ECharts): void {
  const group = groupMap[groupName as string]
  if (!group) {
    return
  }

  const idx = group.indexOf(echart)
  if (idx >= 0) {
    group.splice(idx, 1)
  }
}

export function connect(echart: echarts.ECharts, group: echarts.ECharts[]) {
  echart.on('updateAxisPointer', function (params: any) {
    const payload = (echart as any).makeActionFromEvent(params)

    const axesInfo = payload.axesInfo || []
    for (let i = axesInfo.length - 1; i >= 0; i--) {
      if (axesInfo[i].axisDim === 'y') {
        axesInfo.splice(i, 1)
      }
    }

    for (let c of group) {
      if (c === echart) {
        continue
      }

      delete payload.axesInfo
      ;(c as any).dispatchAction(payload, true)
    }
  })
}
</script>

<style lang="scss">
.chart-tooltip {
  font-size: 0.85rem;

  p {
    margin-bottom: 2px;
    font-weight: 600;
  }

  table {
    width: 100%;
    border-collapse: collapse;

    tr.highlighted {
      font-weight: 600;
    }

    td {
      &:first-child {
        padding-top: 2px;
        padding-right: 2px;
      }

      &:last-child {
        padding-left: 15px;
        text-align: right;
        font-weight: 600;
      }
    }
  }
}
</style>
