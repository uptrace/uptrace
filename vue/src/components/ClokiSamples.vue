<template>
  <v-simple-table>
    <tbody>
      <tr v-for="(sample, i) in items" :key="i">
        <td class="text-no-wrap"><XDate :date="sample.time" format="full" /></td>
        <td>{{ sample.string }}</td>
      </tr>
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Composables
import { Sample } from '@/use/cloki'

export default defineComponent({
  name: 'ClokiSamples',

  props: {
    items: {
      type: Array as PropType<Sample[]>,
      required: true,
    },
  },
})
</script>
