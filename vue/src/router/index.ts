import Vue from 'vue'
import VueRouter, { RouteConfig, NavigationGuard } from 'vue-router'

// Utilities
import { xkey } from '@/models/otelattr'

// Composables
import { useUser } from '@/use/org'
import { buildGroupBy } from '@/use/uql'
import { System } from '@/use/systems'

import Overview from '@/tracing/views/Overview.vue'
import SystemOverview from '@/tracing/views/SystemOverview.vue'
import SlowestGroups from '@/tracing/views/SlowestGroups.vue'
import SystemGroupList from '@/tracing/views/SystemGroupList.vue'
import AttrOverview from '@/tracing/views/AttrOverview.vue'

import TracingHelp from '@/tracing/views/Help.vue'
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
      query: buildGroupBy(xkey.spanGroupId),
      spanListRoute: 'SpanList',
      groupListRoute: 'SpanGroupList',
    },
    children: [
      {
        name: 'SpanGroupList',
        path: 'groups',
        component: GroupList,
      },
      {
        name: 'SpanList',
        path: 'spans',
        component: SpanList,
      },
      {
        path: '',
        redirect: { name: 'SpanGroupList' },
      },
    ],
  },

  {
    path: '/logs/:projectId(\\d+)',
    component: Tracing,
    props: {
      query: `group by ${xkey.spanGroupId} | ${xkey.spanCountPerMin}`,
      spanListRoute: 'LogList',
      groupListRoute: 'LogGroupList',
      systemsFilter: (items: System[]) => {
        items = items.filter((item: System) => item.system.startsWith('log:'))

        if (items.length === 0) {
          items.push({
            system: 'log:all',
            isEvent: true,
          } as System)
        }

        return items
      },
      showLogql: true,
    },
    children: [
      {
        name: 'LogList',
        path: 'spans',
        component: SpanList,
      },
      {
        name: 'LogGroupList',
        path: 'groups',
        component: GroupList,
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
