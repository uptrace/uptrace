<template>
  <v-container fluid class="fill-height grey lighten-5">
    <v-row>
      <v-col class="text-center">
        <h1 v-if="title" class="text-h1">{{ title }}</h1>
        <h2 class="mt-4 text-h5">{{ message }}</h2>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed } from '@vue/composition-api'

export default defineComponent({
  name: 'TraceError',
  props: {
    error: {
      type: Error,
      required: true,
    },
  },

  setup(props) {
    const error = computed((): any => {
      return props.error
    })

    const title = computed(() => {
      return error.value?.response?.data?.status ?? ''
    })

    const message = computed(() => {
      return error.value?.response?.data?.message ?? String(props.error)
    })

    return { title, message }
  },
})
</script>

<style lang="scss" scoped></style>
