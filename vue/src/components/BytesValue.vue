<script lang="ts">
import Vue from 'vue'
import { bytes, numVerbose } from '@/util/fmt'

export default Vue.component('BytesValue', {
  functional: true,
  props: {
    value: {
      type: Number,
      default: undefined,
    },
  },

  render(h, { props, data }) {
    data.attrs = {
      ...data.attrs,
      title: numVerbose(props.value),
    }
    return h('span', data, bytes(props.value))
  },
})
</script>

<style lang="scss" scoped></style>
