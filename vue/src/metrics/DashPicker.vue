<template>
  <v-autocomplete
    :value="value"
    :items="filteredItems"
    item-value="id"
    item-text="name"
    :search-input.sync="searchInput"
    no-filter
    placeholder="dashboard"
    hide-details
    dense
    outlined
    auto-select-first
    background-color="light-blue lighten-5"
    @change="onChange"
  >
  </v-autocomplete>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch, watchEffect, PropType } from 'vue'

// Composables
import { useRouter, useRoute } from '@/use/router'
import { useStorage } from '@/use/local-storage'
import { Dashboard } from '@/metrics/use-dashboards'

export default defineComponent({
  name: 'DashPicker',

  props: {
    value: {
      type: [String, Number],
      default: undefined,
    },
    items: {
      type: Array as PropType<Dashboard[]>,
      required: true,
    },
  },

  setup(props) {
    const { router } = useRouter()
    const route = useRoute()
    const searchInput = shallowRef('')

    const { item: lastDashId } = useStorage<string | number>(
      computed(() => {
        const projectId = route.value.params.projectId ?? 0
        return `last-dashboard:${projectId}`
      }),
    )

    const filteredItems = computed(() => {
      if (!searchInput.value) {
        return props.items
      }

      const index = props.items.findIndex((item) => item.name === searchInput.value)
      if (index >= 0) {
        return props.items
      }

      return fuzzyFilter(props.items, searchInput.value, { key: 'name' })
    })

    watchEffect(
      () => {
        if (!props.items.length) {
          return
        }

        if (!props.value) {
          redirectToLast()
          return
        }

        const index = props.items.findIndex((d) => d.id === props.value)
        if (index === -1) {
          redirectToLast()
          return
        }
      },
      { flush: 'post' },
    )

    watch(
      () => props.value,
      (dashId) => {
        if (dashId) {
          lastDashId.value = dashId
        }
      },
      { immediate: true },
    )

    function onChange(dashId: string) {
      const found = props.items.find((d) => d.id === dashId)
      if (found) {
        redirectTo(found)
      }
    }

    function redirectToLast() {
      let found = props.items.find((d) => d.id === lastDashId.value)
      if (!found) {
        found = props.items[0]
      }
      redirectTo(found)
    }

    function redirectTo(dash: Dashboard) {
      router.push({ name: 'MetricsDashShow', params: { dashId: dash.id } })
    }

    return { searchInput, filteredItems, onChange }
  },
})
</script>

<style lang="scss" scoped></style>
