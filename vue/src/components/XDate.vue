<script lang="ts">
import Vue from 'vue'
import { toDate, datetime } from '@/util/date'

export default Vue.component('XDate', {
  functional: true,
  props: {
    date: {
      type: [String, Date],
      required: true,
    },
  },

  render(h, { props, data }) {
    const dt = toDate(props.date)
    return h('span', data, datetime(dt))
  },
})
</script>

<style lang="scss" scoped></style>
