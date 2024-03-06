import Vue from 'vue'

import PageToolbar from '@/components/PageToolbar.vue'
import XPlaceholder from '@/components/XPlaceholder.vue'
import XPagination from '@/components/XPagination.vue'
import NumValue from '@/components/NumValue.vue'
import PctValue from '@/components/PctValue.vue'
import AnyValue from '@/components/AnyValue.vue'
import DateValue from '@/components/DateValue.vue'
import DurationValue from '@/components/DurationValue.vue'
import BytesValue from '@/components/BytesValue.vue'
import PrismCode from '@/components/PrismCode.vue'
import PagedSpansCardLazy from '@/tracing/PagedSpansCardLazy.vue'
import PagedGroupsCard from '@/tracing/PagedGroupsCard.vue'

Vue.component('PageToolbar', PageToolbar)
Vue.component('XPlaceholder', XPlaceholder)
Vue.component('XPagination', XPagination)
Vue.component('NumValue', NumValue)
Vue.component('PctValue', PctValue)
Vue.component('AnyValue', AnyValue)
Vue.component('DateValue', DateValue)
Vue.component('DurationValue', DurationValue)
Vue.component('BytesValue', BytesValue)
Vue.component('PrismCode', PrismCode)
Vue.component('PagedSpansCardLazy', PagedSpansCardLazy)
Vue.component('PagedGroupsCard', PagedGroupsCard)

Vue.mixin({
  data() {
    return {
      publicPath: process.env.BASE_URL,
    }
  },
  computed: {
    console: () => console,
  },
})
