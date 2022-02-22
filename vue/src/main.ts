import Vue from 'vue'

import VueCompositionApi from '@vue/composition-api'
Vue.use(VueCompositionApi)

import App from '@/App.vue'
import router from '@/router'
import '@/styles/index.scss'

import '@/plugins/directives'
import '@/plugins/frag'
import '@/plugins/axios'
import '@/plugins/prism'
import '@/plugins/global'
import vuetify from '@/plugins/vuetify'

Vue.config.productionTip = false

new Vue({
  router,
  vuetify,
  render: (h) => h(App),
}).$mount('#app')
