<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ attrs, on }">
      <v-btn
        dark
        class="blue darken-1 elevation-5"
        style="text-transform: none"
        v-bind="attrs"
        v-on="on"
      >
        <span class="px-4">{{
          dashboards.active ? dashboards.active.name : 'Choose dashboard'
        }}</span>
        <v-icon right size="24">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <DashTree :tree="dashboards.tree" @change="menu = false" />
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDashboards } from '@/metrics/use-dashboards'

// Components
import DashTree from '@/metrics/DashTree.vue'

export default defineComponent({
  name: 'DashPicker',
  components: { DashTree },

  props: {
    dashboards: {
      type: Object as PropType<UseDashboards>,
      required: true,
    },
    maxHeight: {
      type: Number,
      default: 420,
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const { router, route } = useRouter()

    watch(
      () => props.dashboards.items,
      (dashboards) => {
        if (route.value.params.dashId) {
          return
        }
        if (!dashboards.length) {
          return
        }

        const found = dashboards[0]
        router.replace({ name: 'MetricsDashShow', params: { dashId: found.id } })
      },
      { immediate: true },
    )

    return { menu }
  },
})
</script>

<style lang="scss" scoped></style>
