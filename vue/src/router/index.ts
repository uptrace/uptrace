import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'

import Overview from '@/views/Overview.vue'
import Tracing from '@/views/Tracing.vue'
import TraceShow from '@/views/TraceShow.vue'
import SpanShow from '@/views/SpanShow.vue'
import Help from '@/views/Help.vue'

Vue.use(VueRouter)

const routes: Array<RouteConfig> = [
  {
    name: 'Home',
    path: '/',
    redirect: { name: 'GroupList' },
  },
  {
    name: 'Overview',
    path: '/overview',
    component: Overview,
  },
  {
    name: 'GroupList',
    path: '/explore',
    component: Tracing,
  },
  {
    name: 'TraceShow',
    path: '/traces/:traceId',
    component: TraceShow,
  },
  {
    name: 'SpanShow',
    path: '/traces/:traceId/:spanId',
    component: SpanShow,
  },
  {
    name: 'Help',
    path: '/help',
    component: Help,
  },
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
})

export default router
