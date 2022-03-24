import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'

// Composables
import { useUser } from '@/use/org'

import Overview from '@/views/Overview.vue'
import SystemOverview from '@/components/SystemOverview.vue'
import ServiceOverview from '@/components/ServiceOverview.vue'
import HostOverview from '@/components/HostOverview.vue'
import SlowestGroups from '@/components/SlowestGroups.vue'
import SystemGroupList from '@/components/SystemGroupList.vue'

import Tracing from '@/views/Tracing.vue'
import GroupList from '@/views/GroupList.vue'
import SpanList from '@/views/SpanList.vue'

import TraceShow from '@/views/TraceShow.vue'
import TraceFind from '@/views/TraceFind.vue'
import SpanShow from '@/views/SpanShow.vue'

import Login from '@/views/Login.vue'
import Help from '@/views/Help.vue'

Vue.use(VueRouter)

const routes: Array<RouteConfig> = [
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
    name: 'Help',
    path: '/help/:projectId(\\d+)',
    component: Help,
  },
  {
    name: 'ProjectCreate',
    path: '/help/projects',
    component: Help,
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
    path: '/services/:projectId(\\d+)',
    component: Overview,
    children: [
      {
        name: 'ServiceOverview',
        path: '',
        component: ServiceOverview,
      },
    ],
  },
  {
    path: '/hosts/:projectId(\\d+)',
    component: Overview,
    children: [
      {
        name: 'HostOverview',
        path: '',
        component: HostOverview,
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
    path: '/overview/:projectId(\\d+)/:system',
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
    path: '/explore/:projectId(\\d+)',
    component: Tracing,
    children: [
      {
        name: 'GroupList',
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
        redirect: { name: 'GroupList' },
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
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
})

export default router
