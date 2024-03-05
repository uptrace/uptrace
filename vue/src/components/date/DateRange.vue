<script lang="ts">
import { format } from 'date-fns'
import { defineComponent, h } from 'vue'
import { toDate } from '@/util/fmt/date'

export default defineComponent({
  name: 'DateRange',

  props: {
    start: {
      type: [Number, String, Date],
      required: true,
    },
    end: {
      type: [Number, String, Date],
      required: true,
    },
  },

  setup(props) {
    return () => {
      if (!props.start || !props.end) {
        return h('span', '')
      }

      const start = toDate(props.start)
      const end = toDate(props.end)

      if (start.getHours() === end.getHours() && start.getMinutes() === end.getMinutes()) {
        const endPattern = start.getMonth() === end.getMonth() ? 'dd HH:mm' : 'LLL dd HH:mm'
        return h('span', [h('span', `${format(start, 'LLL dd')} - ${format(end, endPattern)}`)])
      }

      if (start.getMonth() === end.getMonth() && start.getDate() == end.getDate()) {
        return h('span', [h('span', `${format(start, 'LLL dd HH:mm')} - ${format(end, 'HH:mm')}`)])
      }

      const fullFormat = 'LLL d HH:mm'
      return h('span', [
        h('span', format(start, fullFormat)),
        h('span', ' - '),
        h('span', format(end, fullFormat)),
      ])
    }
  },
})
</script>

<style lang="scss" scoped></style>
