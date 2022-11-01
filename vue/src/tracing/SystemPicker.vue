<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn style="text-transform: none" v-bind="attrs" v-on="on">
        <span>{{ systems.activeSystem || 'Choose system' }}</span>
        <v-icon right size="24">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <SystemList
      :date-range="dateRange"
      :items="systemsTree"
      :max-height="maxHeight"
      @click:item="menu = false"
    />
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watchEffect, PropType } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { buildSystemsTree, UseSystems, System } from '@/tracing/use-systems'

// Components
import SystemList from '@/tracing/SystemList.vue'

// Utilities
import { SystemName } from '@/models/otelattr'

export default defineComponent({
  name: 'SystemPicker',
  components: { SystemList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    items: {
      type: Array as PropType<System[]>,
      required: true,
    },
    allSystem: {
      type: String,
      required: true,
    },
    outlined: {
      type: Boolean,
      default: false,
    },
    maxHeight: {
      type: Number,
      default: 420,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    const attrs = computed(() => {
      if (props.outlined) {
        return { outlined: true }
      }
      return { dark: true, class: 'blue darken-1 elevation-5' }
    })

    const items = computed(() => {
      const items = [...props.items]
      items.unshift({
        projectId: items[0].projectId,
        system: props.allSystem,
        text: SystemName.all,
        isEvent: true,
        count: 0,
        rate: 0,
        errorCount: 0,
        errorPct: 0,
        dummy: true,
      })
      return items
    })

    const systemsTree = computed(() => {
      const tree = buildSystemsTree(items.value)
      if (tree.length === 1 && tree[0].children) {
        return items.value
      }
      return tree
    })

    const activeItem = computed(() => {
      return (
        items.value.find((item: System) => item.system === props.systems.activeSystem) ||
        items.value[0]
      )
    })

    useRouteQuery().sync({
      fromQuery(query) {
        if (typeof query.system === 'string') {
          props.systems.change(query.system)
        } else {
          props.systems.reset()
        }
      },
      toQuery() {
        if (props.systems.activeSystem) {
          return { system: props.systems.activeSystem }
        }
      },
    })

    watchEffect(
      () => {
        if (props.systems.activeSystem) {
          return
        }
        if (items.value.length) {
          props.systems.change(items.value[0].system)
        }
      },
      { flush: 'post' },
    )

    return {
      menu,
      activeItem,
      attrs,
      systemsTree,
    }
  },
})
</script>

<style lang="scss" scoped></style>
