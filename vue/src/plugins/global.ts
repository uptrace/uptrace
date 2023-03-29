import Vue from 'vue'

import PageToolbar from '@/components/PageToolbar.vue'
import XPlaceholder from '@/components/XPlaceholder.vue'
import XPagination from '@/components/XPagination.vue'
import XNum from '@/components/XNum.vue'
import XPct from '@/components/XPct.vue'
import AnyValue from '@/components/AnyValue.vue'
import XDate from '@/components/XDate.vue'
import XDuration from '@/components/XDuration.vue'
import PrismCode from '@/components/PrismCode.vue'
import DashGaugeCard from '@/metrics/gauge/DashGaugeCard.vue'

Vue.component('PageToolbar', PageToolbar)
Vue.component('XPlaceholder', XPlaceholder)
Vue.component('XPagination', XPagination)
Vue.component('XNum', XNum)
Vue.component('XPct', XPct)
Vue.component('AnyValue', AnyValue)
Vue.component('XDate', XDate)
Vue.component('XDuration', XDuration)
Vue.component('PrismCode', PrismCode)
Vue.component('DashGaugeCard', DashGaugeCard)

Vue.mixin({
  computed: {
    console: () => console,
  },
})
