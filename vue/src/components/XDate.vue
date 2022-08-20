<script lang="ts">
import Vue from 'vue'
import {
  toDate,
  date,
  dateShort,
  time,
  datetime,
  datetimeShort,
  datetimeFull,
  relative,
  fromNow,
} from '@/util/fmt/date'

export default Vue.component('XDate', {
  functional: true,
  props: {
    date: {
      type: [String, Date],
      required: true,
    },
    format: {
      type: String,
      default: '',
    },
  },

  render(h, { props, data }) {
    const fmt = formatter(props.format)

    if (!data.attrs) {
      data.attrs = {}
    }

    const dt = toDate(props.date)

    if (fmt !== datetimeFull) {
      data.attrs.title = datetimeFull(dt)
    }

    return h('span', data, fmt(dt))
  },
})

function formatter(format: string) {
  switch (format) {
    case '':
      return datetime
    case 'time':
      return time
    case 'date':
      return date
    case 'dateShort':
      return dateShort
    case 'short':
      return datetimeShort
    case 'full':
      return datetimeFull
    case 'relative':
      return relative
    case 'from-now':
      return fromNow
    default:
      throw new Error(`unknown format=${format}`)
  }
}
</script>

<style lang="scss" scoped></style>
