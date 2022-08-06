<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn style="text-transform: none" v-bind="attrs" v-on="on">
        <span class="px-4">{{ systems.activeSystem || 'Choose system' }}</span>
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
import { buildSystemsTree, UseSystems, System } from '@/use/systems'

// Components
import SystemList from '@/components/SystemList.vue'

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

    const systemsTree = computed(() => {
      const tree = buildSystemsTree(props.items)
      if (tree.length === 1 && tree[0].children) {
        return props.items
      }
      return tree
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

    watchEffect(() => {
      if (props.systems.activeSystem) {
        return
      }
      if (props.items.length) {
        props.systems.change(props.items[0].system)
      }
    })

    return { menu, attrs, systemsTree }
  },
})
</script>

<style lang="scss" scoped></style>
