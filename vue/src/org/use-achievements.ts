import { proxyRefs, computed, ComputedRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'
import { useForceReload } from '@/use/force-reload'
import { Project } from '@/org/use-projects'

export enum AchievName {
  ConfigureTracing = 'configure-tracing',
  ConfigureMetrics = 'configure-metrics',
  InstallCollector = 'install-collector',
  CreateMetricMonitor = 'create-metric-monitor',
}

export interface AchievementData {
  id: string
  userId: number
  projectId: number
  name: AchievName
}

interface Achievement {
  name: AchievName
  title: string
  subtitle: string
  data?: AchievementData
  attrs: ListItemAttrs
}

interface ListItemAttrs {
  to?: Record<string, any>
  href?: string
  target?: string
}

export type UseAchievements = ReturnType<typeof useAchievements>

export const useAchievements = function (project: ComputedRef<Project | undefined>) {
  const { route } = useRouter()
  const { forceReloadParams } = useForceReload()

  const allAchievements = computed((): Achievement[] => {
    const items: Achievement[] = [
      {
        name: AchievName.ConfigureTracing,
        title: 'Start sending traces to Uptrace',
        subtitle: 'Configure OpenTelemetry to export spans to Uptrace',
        attrs: {
          to: { name: 'TracingHelp' },
        },
      },
      {
        name: AchievName.ConfigureMetrics,
        title: 'Start sending metrics to Uptrace',
        subtitle: 'Configure OpenTelemetry to export metrics to Uptrace',
        attrs: {
          to: { name: 'MetricsHelp' },
        },
      },
      {
        name: AchievName.InstallCollector,
        title: 'Install OpenTelemetry Collector',
        subtitle: 'Monitor infrastructure metrics with Otel Collector',
        attrs: {
          href: 'https://uptrace.dev/get/ingest/collector.html',
          target: '_blank',
        },
      },
      {
        name: AchievName.CreateMetricMonitor,
        title: 'Monitor spans and metrics',
        subtitle: 'Get notified to resolve incidents in time',
        attrs: {
          to: {
            name: 'MonitorList',
          },
        },
      },
    ]

    return items
  })

  const { data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/projects/${projectId}/achievements`,
      params: forceReloadParams.value,
    }
  })

  const completed = computed((): AchievementData[] => {
    return data.value?.achievements ?? []
  })

  const achievements = computed((): Achievement[] => {
    return allAchievements.value.map((achv) => {
      const data = completed.value.find((item) => item.name === achv.name)
      if (data) {
        achv = mergeAchievement(achv, data)
      }
      return achv
    })
  })

  function isCompleted(achievement: Achievement) {
    return Boolean(achievement.data)
  }

  return proxyRefs({
    reload,

    items: achievements,
    completed,
    isCompleted,
  })
}

function mergeAchievement(achv: Achievement, other: AchievementData): Achievement {
  return {
    ...achv,
    data: other,
  }
}
