import Vue from 'vue'
import VueRouter, { RouteConfig, NavigationGuard } from 'vue-router'

// Utilities
import { xkey, xsys, isEventSystem } from '@/models/otelattr'

// Composables
import { useUser } from '@/use/org'
import { exploreAttr } from '@/use/uql'
import { System } from '@/use/systems'

import Overview from '@/tracing/views/Overview.vue'
import SystemOverview from '@/tracing/views/SystemOverview.vue'
import SlowestGroups from '@/tracing/views/SlowestGroups.vue'
import SystemGroupList from '@/tracing/views/SystemGroupList.vue'
import AttrOverview from '@/tracing/views/AttrOverview.vue'

import TracingHelp from '@/tracing/views/Help.vue'
import ProjectSettings from '@/org/views/ProjectSettings.vue'
import Tracing from '@/tracing/views/Tracing.vue'
import GroupList from '@/tracing/views/GroupList.vue'
import SpanList from '@/tracing/views/SpanList.vue'

import TraceShow from '@/tracing/views/TraceShow.vue'
import TraceFind from '@/tracing/views/TraceFind.vue'
import SpanShow from '@/tracing/views/SpanShow.vue'

import MetricsLayout from '@/metrics/views/Layout.vue'
import MetricsDash from '@/metrics/views/Dashboard.vue'
import MetricsExplore from '@/metrics/views/Explore.vue'
import MetricsHelp from '@/metrics/views/Help.vue'

import Login from '@/views/Login.vue'

Vue.use(VueRouter)

const routes: RouteConfig[] = [
  {
    name: 'Home',
    path: '/',
    beforeEnter: async (_to, _from, next) => {
      const user = useUser()
      await user.getOrLoad()

      const first = user.projects[0]
      if (first) {
        next({
          name: 'Overview',
          params: { projectId: String(first.id) },
        })
        return
      }

      next({ name: 'ProjectCreate' })
    },
  },

  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    name: 'TracingHelp',
    path: '/help/:projectId(\\d+)',
    component: TracingHelp,
  },
  {
    name: 'ProjectCreate',
    path: '/help/projects',
    component: TracingHelp,
  },
  {
    name: 'ProjectSettings',
    path: '/projects/:projectId(\\d+)',
    component: ProjectSettings,
  },

  {
    path: '/overview/:projectId(\\d+)',
    component: Overview,
    children: [
      {
        name: 'Overview',
        path: '',
        component: SystemOverview,
      },
    ],
  },
  {
    path: '/slowest-groups/:projectId(\\d+)',
    component: Overview,
    children: [
      {
        name: 'SlowestGroups',
        path: '',
        component: SlowestGroups,
      },
    ],
  },
  {
    path: '/systems/:projectId(\\d+)/:system',
    component: Overview,
    children: [
      {
        name: 'SystemGroupList',
        path: '',
        component: SystemGroupList,
      },
    ],
  },
  {
    path: '/attributes/:projectId(\\d+)/:attr',
    component: Overview,
    children: [
      {
        name: 'AttrOverview',
        path: '',
        component: AttrOverview,
      },
    ],
  },

  {
    path: '/explore/:projectId(\\d+)',
    component: Tracing,
    props: {
      query: exploreAttr(xkey.spanGroupId),
      spanListRoute: 'SpanList',
      groupListRoute: 'SpanGroupList',
      systemsFilter: (items: System[]) => {
        return items.filter((item: System) => !isEventSystem(item.system))
      },
      allSystem: xsys.allSpans,
    },
    children: [
      {
        name: 'SpanGroupList',
        path: '',
        component: GroupList,
      },
      {
        name: 'SpanList',
        path: 'spans',
        component: SpanList,
      },
    ],
  },

  {
    path: '/events/:projectId(\\d+)',
    component: Tracing,
    props: {
      query: exploreAttr(xkey.spanGroupId, true),
      spanListRoute: 'EventList',
      groupListRoute: 'EventGroupList',
      systemsFilter: (items: System[]) => {
        return items.filter((item: System) => isEventSystem(item.system))
      },
      allSystem: xsys.allEvents,
      showLogql: true,
    },
    children: [
      {
        name: 'EventGroupList',
        path: '',
        component: GroupList,
      },
      {
        name: 'EventList',
        path: 'spans',
        component: SpanList,
      },
    ],
  },

  {
    name: 'TraceShow',
    path: '/traces/:projectId(\\d+)/:traceId',
    component: TraceShow,
  },
  {
    name: 'SpanShow',
    path: '/traces/:projectId(\\d+)/:traceId/:spanId',
    component: SpanShow,
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
        name: 'MetricsDashList',
        component: MetricsDash,
      },
      {
        path: ':dashId(\\d+)',
        name: 'MetricsDashShow',
        component: MetricsDash,
      },
      {
        path: 'explore',
        name: 'MetricsExplore',
        component: MetricsExplore,
      },
    ],
  },
  {
    path: '/metrics/:projectId(\\d+)/help',
    name: 'MetricsHelp',
    component: MetricsHelp,
  },
  {
    path: '/metrics',
    beforeEnter: redirectToProject('MetricsDashList'),
  },
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
})

function redirectToProject(routeName: string): NavigationGuard {
  return async (_to, _from, next) => {
    const user = useUser()
    await user.getOrLoad()

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
