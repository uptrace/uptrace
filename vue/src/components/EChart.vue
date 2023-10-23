<template>
  <div ref="container" v-element-resize @resize="onResize">
    <v-sheet :width="width" :height="height" class="mx-auto echart"></v-sheet>
    <v-menu
      v-model="popover.menu"
      :position-x="popover.x"
      :position-y="popover.y"
      absolute
      :close-on-content-click="false"
      z-index="999"
    >
      <v-card v-if="popover.annotation">
        <v-container fluid class="py-4">
          <v-row align="center" class="text-subtitle-1">
            <v-col cols="auto" class="font-weight-medium">
              {{ popover.annotation.name }}
            </v-col>
            <v-col cols="auto" class="text-body-2 text--secondary">
              <DateValue :value="popover.annotation.createdAt" format="relative" />
            </v-col>
            <v-col cols="auto">
              <v-btn
                small
                text
                color="primary"
                :to="{ name: 'AnnotationShow', params: { annotationId: popover.annotation.id } }"
                >Edit</v-btn
              >
            </v-col>
          </v-row>

          <v-row v-if="Object.keys(popover.annotation.attrs).length" dense>
            <v-col>
              <AnnotationAttrs :attrs="popover.annotation.attrs" />
            </v-col>
          </v-row>

          <v-row v-if="popover.annotation.description">
            <v-col>
              <div v-html="popover.descriptionMarkdown"></div>
            </v-col>
          </v-row>
        </v-container>
      </v-card>
    </v-menu>
  </div>
</template>

<script lang="ts">
import markdownit from 'markdown-it'
import { cloneDeep, debounce } from 'lodash-es'
import { init as initChart, ECharts } from 'echarts'
import colors from 'vuetify/lib/util/colors'
import {
  defineComponent,
  shallowRef,
  computed,
  proxyRefs,
  watch,
  onMounted,
  onUnmounted,
  PropType,
} from 'vue'

// Composables
import { Annotation } from '@/org/use-annotations'

// Components
import AnnotationAttrs from '@/alerting/AnnotationAttrs.vue'

// Utilities
import type { EChartsOption } from '@/util/chart'

type GroupName = string | symbol

export interface EChartProps {
  width?: number | string
  height: number | string
  option: EChartsOption
}

export default defineComponent({
  name: 'EChart',
  components: { AnnotationAttrs },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Object as PropType<ECharts>,
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
      type: Object as PropType<EChartsOption>,
      default: undefined,
    },
    group: {
      type: [String, Symbol] as PropType<GroupName>,
      default: undefined,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props, ctx) {
    let echart: ECharts
    const container = shallowRef<HTMLElement>()

    const config = computed(() => {
      if (!props.option) {
        return undefined
      }
      const conf = cloneDeep(props.option)
      for (let ann of props.annotations) {
        addAnnotation(conf, ann)
      }
      return conf
    })

    function init() {
      if (echart) {
        return
      }

      const div = container.value!.getElementsByClassName('echart')[0] as HTMLDivElement
      echart = initChart(div)

      ctx.emit('input', echart)

      initAnnotations(echart)

      if (props.group) {
        register(props.group, echart)
      }
    }

    const popover = usePopover()
    function initAnnotations(echart: ECharts) {
      const dom = echart.getDom()
      echart.on('click', function (params: any) {
        const annId = parseInt(params.seriesId, 10)
        if (!annId) {
          return
        }

        const found = props.annotations.find((ann) => ann.id == annId)
        if (!found) {
          return
        }

        const event = params.event.event
        popover.annotation = found
        popover.x = event.clientX
        popover.y = dom.getBoundingClientRect().top + dom.clientHeight - 25
        popover.menu = true
      })
    }

    const setOption = debounce((option: EChartsOption) => {
      if (echart.isDisposed()) {
        return
      }

      echart.setOption(option, { notMerge: true, lazyUpdate: true, silent: true })
    }, 50)

    onMounted(() => {
      init()

      watch(
        config,
        (config) => {
          if (config) {
            setOption(config)
          }
        },
        { immediate: true },
      )

      watch(
        () => props.loading,
        (loading) => {
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

    const onResize = debounce(() => {
      if (echart) {
        echart.resize()
      }
    }, 50)

    return { container, onResize, popover }
  },
})

//------------------------------------------------------------------------------

const groupMap: Record<GroupName, ECharts[]> = {}

function register(groupName: GroupName, echart: ECharts): void {
  let group = groupMap[groupName as string]
  if (!group) {
    group = []
    groupMap[groupName as string] = group
  }
  group.push(echart)
  connect(echart, group)
}

function unregister(groupName: GroupName, echart: ECharts): void {
  const group = groupMap[groupName as string]
  if (!group) {
    return
  }

  const idx = group.indexOf(echart)
  if (idx >= 0) {
    group.splice(idx, 1)
  }
}

export function connect(echart: ECharts, group: ECharts[]) {
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

function addAnnotation(conf: EChartsOption, ann: Annotation) {
  let selected: Record<string, boolean> | undefined = undefined
  if (conf.legend && conf.legend.length) {
    const legend = conf.legend[0]
    if (legend.selected) {
      selected = legend.selected
    }
  }

  let min = 0
  let max = Number.NaN

  for (let ds of conf.dataset) {
    const source = ds.source as Record<string, number[]>
    for (let key in source) {
      if (key === 'time') {
        continue
      }
      if (selected && !selected[key]) {
        continue
      }

      const value = source[key]
      if (!Array.isArray(value)) {
        continue
      }

      const dsMin = Math.min(...value)
      if (Number.isNaN(min) || dsMin < min) {
        min = dsMin
      }

      const dsMax = Math.max(...value)
      if (Number.isNaN(max) || dsMax > max) {
        max = dsMax
      }
    }
  }

  if (Number.isNaN(max) || max === min) {
    max = min + 1
  }

  const time = ann.createdAt
  conf.series.push({
    id: ann.id,
    name: '_',
    type: 'line',
    data: [{ value: [time, min], symbol: 'square' }, { value: [time, max] }],
    color: ann.color || colors.pink.darken1,
    lineStyle: { width: 2, type: 'dashed' },
    symbol: 'none',
    symbolSize: 10,
    z: 999,
    triggerLineEvent: true,
  })
}

const md = markdownit()

function usePopover() {
  const menu = shallowRef(false)
  const annotation = shallowRef<Annotation>()
  const x = shallowRef(0)
  const y = shallowRef(0)

  const descriptionMarkdown = computed(() => {
    const text = annotation.value?.description ?? ''
    return md.render(text)
  })

  const attrKeys = computed(() => {
    const keys = Object.keys(annotation.value?.attrs ?? {})
    keys.sort()
    return keys
  })

  return proxyRefs({ menu, annotation, x, y, descriptionMarkdown, attrKeys })
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
