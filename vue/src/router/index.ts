import Vue from 'vue'
import VueRouter, { RouteConfig, NavigationGuard } from 'vue-router'

// Composables
import { useUser } from '@/org/use-users'

import NotFoundPage from '@/org/views/NotFoundPage.vue'
import Project from '@/org/views/Project.vue'
import ProjectSettings from '@/org/views/ProjectSettings.vue'
import ProjectDsn from '@/org/views/ProjectDsn.vue'

import Alerting from '@/alerting/views/Alerting.vue'
import AlertList from '@/alerting/views/AlertList.vue'
import AlertShow from '@/alerting/views/AlertShow.vue'
import MonitorList from '@/alerting/views/MonitorList.vue'
import MonitorMetric from '@/alerting/views/MonitorMetric.vue'
import MonitorErrorShow from '@/alerting/views/MonitorErrorShow.vue'
import MonitorErrorNew from '@/alerting/views/MonitorErrorNew.vue'
import ChannelList from '@/alerting/views/ChannelList.vue'
import ChannelShowSlack from '@/alerting/views/ChannelShowSlack.vue'
import ChannelShowTelegram from '@/alerting/views/ChannelShowTelegram.vue'
import ChannelShowWebhook from '@/alerting/views/ChannelShowWebhook.vue'
import ChannelShowAlertmanager from '@/alerting/views/ChannelShowAlertmanager.vue'
import EmailNotifications from '@/alerting/views/EmailNotifications.vue'
import AnnotationList from '@/alerting/views/AnnotationList.vue'
import AnnotationShow from '@/alerting/views/AnnotationShow.vue'

import Overview from '@/tracing/views/Overview.vue'
import OverviewAttr from '@/tracing/views/OverviewAttr.vue'
import OverviewSlowestGroups from '@/tracing/views/OverviewSlowestGroups.vue'
import OverviewGroups from '@/tracing/views/OverviewGroups.vue'
import OverviewServiceGraph from '@/tracing/views/OverviewServiceGraph.vue'

import TracingHelp from '@/tracing/views/Help.vue'
import TracingCheatsheet from '@/tracing/views/Cheatsheet.vue'
import Tracing from '@/tracing/views/Tracing.vue'
import TracingGroups from '@/tracing/views/TracingGroups.vue'
import TracingSpans from '@/tracing/views/TracingSpans.vue'
import TracingTimeseries from '@/tracing/views/TracingTimeseries.vue'

import TraceShow from '@/tracing/views/TraceShow.vue'
import TraceFind from '@/tracing/views/TraceFind.vue'
import TraceSpanShow from '@/tracing/views/TraceSpanShow.vue'

import MetricsLayout from '@/metrics/views/Layout.vue'
import MetricsExplore from '@/metrics/views/Explore.vue'
import MetricsHelp from '@/metrics/views/Help.vue'
import MetricsCheatsheet from '@/metrics/views/Cheatsheet.vue'

import DashboardList from '@/metrics/views/DashboardList.vue'
import Dashboard from '@/metrics/views/Dashboard.vue'
import DashboardLoading from '@/metrics/views/DashboardLoading.vue'
import DashboardTable from '@/metrics/views/DashboardTable.vue'
import DashboardGrid from '@/metrics/views/DashboardGrid.vue'
import DashboardHelp from '@/metrics/views/DashboardHelp.vue'

import Login from '@/views/Login.vue'
import UserProfile from '@/org/views/UserProfile.vue'
import DataUsage from '@/org/views/DataUsage.vue'

Vue.use(VueRouter)

const routes: RouteConfig[] = [
  {
    name: 'Home',
    path: '/',
    beforeEnter: redirectToProject('Overview'),
  },

  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    path: '/profile',
    name: 'UserProfile',
    component: UserProfile,
  },
  {
    path: '/usage',
    name: 'DataUsage',
    component: DataUsage,
  },
  {
    name: 'ProjectCreate',
    path: '/help/projects',
    component: TracingHelp,
  },
  {
    path: '/projects/:projectId(\\d+)',
    component: Project,
    children: [
      {
        name: 'ProjectShow',
        path: '',
        components: { tab: ProjectSettings },
      },
      {
        name: 'ProjectDsn',
        path: 'dsn',
        components: { tab: ProjectDsn },
      },
    ],
  },

  {
    path: '/alerting/:projectId(\\d+)',
    name: 'Alerting',
    component: Alerting,
    redirect: { name: 'AlertList' },
    children: [
      {
        name: 'AlertList',
        path: 'alerts',
        components: { alerting: AlertList },
      },
      {
        name: 'AlertShow',
        path: 'alerts/:alertId(\\d+)',
        components: { alerting: AlertShow },
      },

      {
        name: 'NotifChannelList',
        path: 'channels',
        components: { alerting: ChannelList },
      },

      {
        name: 'MonitorList',
        path: 'monitors',
        components: { alerting: MonitorList },
      },

      {
        name: 'NotifChannelEmail',
        path: 'email',
        components: { alerting: EmailNotifications },
      },

      {
        name: 'AnnotationList',
        path: 'annotations',
        components: { alerting: AnnotationList },
      },
    ],
  },

  {
    name: 'MonitorMetricNew',
    path: '/alerting/:projectId(\\d+)/monitors/new-metric',
    component: MonitorMetric,
  },
  {
    name: 'MonitorMetricShow',
    path: '/alerting/:projectId(\\d+)/monitors/:monitorId(\\d+)/metric',
    component: MonitorMetric,
  },
  {
    name: 'MonitorErrorNew',
    path: '/alerting/:projectId(\\d+)/monitors/new-error',
    component: MonitorErrorNew,
  },
  {
    name: 'MonitorErrorShow',
    path: '/alerting/:projectId(\\d+)/monitors/:monitorId(\\d+)/error',
    component: MonitorErrorShow,
  },

  {
    name: 'NotifChannelShowSlack',
    path: '/alerting/:projectId(\\d+)/channels/slack/:channelId(\\d+)',
    component: ChannelShowSlack,
  },
  {
    name: 'NotifChannelShowTelegram',
    path: '/alerting/:projectId(\\d+)/channels/telegram/:channelId(\\d+)',
    component: ChannelShowTelegram,
  },
  {
    name: 'NotifChannelShowWebhook',
    path: '/alerting/:projectId(\\d+)/channels/webhook/:channelId(\\d+)',
    component: ChannelShowWebhook,
  },
  {
    name: 'NotifChannelShowAlertmanager',
    path: '/alerting/:projectId(\\d+)/channels/alertmanager/:channelId(\\d+)',
    component: ChannelShowAlertmanager,
  },

  {
    path: '/alerts/:projectId(\\d+)/:alertId(\\d+)',
    redirect: { name: 'AlertList' },
  },
  {
    path: '/alerts',
    beforeEnter: redirectToProject('AlertList'),
  },

  {
    name: 'AnnotationShow',
    path: '/alerting/:projectId(\\d+)/annotations/:annotationId(\\d+)',
    component: AnnotationShow,
  },

  {
    name: 'Overview',
    path: '/overview/:projectId(\\d+)',
    component: Overview,
    redirect: { name: 'SystemOverview' },

    children: [
      {
        name: 'SystemOverview',
        path: 'systems',
        components: { overview: OverviewAttr },
      },
      {
        name: 'SystemGroupList',
        path: 'groups/:system',
        components: { overview: OverviewGroups },
      },
      {
        name: 'AttrOverview',
        path: 'attributes/:attr',
        components: { overview: OverviewAttr },
      },
      {
        name: 'SlowestGroups',
        path: 'slowest-groups',
        components: { overview: OverviewSlowestGroups },
      },
      {
        name: 'ServiceGraph',
        path: 'service-graph',
        components: { overview: OverviewServiceGraph },
      },
    ],
  },
  {
    path: '/:projectId(\\d+)',
    redirect: { name: 'Overview' },
  },
  {
    path: '/overview',
    beforeEnter: redirectToProject('SystemOverview'),
  },

  {
    name: 'TracingHelp',
    path: '/spans/:projectId(\\d+)/help',
    component: TracingHelp,
  },
  {
    path: '/spans/:projectId(\\d+)/cheatsheet',
    name: 'TracingCheatsheet',
    component: TracingCheatsheet,
  },
  {
    path: '/spans/:projectId(\\d+)',
    component: Tracing,
    children: [
      {
        name: 'SpanGroupList',
        path: '',
        components: { tracing: TracingGroups },
      },
      {
        name: 'SpanList',
        path: 'items',
        components: { tracing: TracingSpans },
      },
      {
        name: 'SpanTimeseries',
        path: 'timeseries',
        components: { tracing: TracingTimeseries },
      },
    ],
  },
  {
    path: '/spans',
    beforeEnter: redirectToProject('SpanGroupList'),
  },

  {
    path: '/events',
    beforeEnter: redirectToProject('SpanGroupList'),
  },

  {
    name: 'TraceShow',
    path: '/traces/:projectId(\\d+)/:traceId',
    component: TraceShow,
  },
  {
    name: 'TraceSpanShow',
    path: '/traces/:projectId(\\d+)/:traceId/:spanId',
    component: TraceSpanShow,
  },
  {
    path: '/traces/:traceId',
    name: 'TraceFind',
    component: TraceFind,
  },

  {
    path: '/metrics/:projectId(\\d+)',
    component: MetricsLayout,
    children: [
      {
        path: '',
        name: 'DashboardList',
        components: { metrics: DashboardList },
      },
      {
        path: ':dashId(\\d+)',
        components: { metrics: Dashboard },
        children: [
          {
            path: '',
            name: 'DashboardShow',
            components: { tab: DashboardLoading },
          },
          {
            path: 'table',
            name: 'DashboardTable',
            components: { tab: DashboardTable },
          },
          {
            path: 'grid',
            name: 'DashboardGrid',
            components: { tab: DashboardGrid },
          },
          {
            path: 'help',
            name: 'DashboardHelp',
            components: { tab: DashboardHelp },
          },
        ],
      },
      {
        path: 'explore',
        name: 'MetricsExplore',
        components: { metrics: MetricsExplore },
      },
    ],
  },
  {
    path: '/metrics/:projectId(\\d+)/help',
    name: 'MetricsHelp',
    component: MetricsHelp,
  },
  {
    path: '/metrics/:projectId(\\d+)/cheatsheet',
    name: 'MetricsCheatsheet',
    component: MetricsCheatsheet,
  },
  {
    path: '/metrics',
    beforeEnter: redirectToProject('DashboardList'),
  },

  { path: '*', component: NotFoundPage },
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      return {
        selector: to.hash,
      }
    }
    return savedPosition
  },
})

function redirectToProject(routeName: string): NavigationGuard {
  return async (_to, _from, next) => {
    const user = useUser()
    await user.getOrLoad()

    for (let p of user.projects) {
      if (p.id === user.lastProjectId) {
        next({ name: routeName, params: { projectId: String(p.id) } })
        return
      }
    }

    const first = user.projects[0]
    if (first) {
      next({
        name: routeName,
        params: { projectId: String(first.id) },
      })
      return
    }

    next({ name: 'ProjectCreate' })
  }
}

export default router
