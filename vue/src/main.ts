import Vue from 'vue'

import App from '@/App.vue'
import router from '@/router'

import 'roboto-fontface/css/roboto/roboto-fontface.css'
import '@mdi/font/css/materialdesignicons.min.css'
import '@/styles/index.scss'

import '@/plugins/directives'
import '@/plugins/frag'
import '@/plugins/axios'
import '@/plugins/prism'
import '@/plugins/global'
import vuetify from '@/plugins/vuetify'
import '@/plugins/portal'

Vue.config.productionTip = false

new Vue({
  router,
  vuetify,
  render: (h) => h(App),
}).$mount('#app')
