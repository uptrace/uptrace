<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on }">
      <v-btn style="text-transform: none" v-bind="attrs" v-on="on">
        <span class="px-4">{{ systems.activeValue || 'Choose system' }}</span>
        <v-icon right size="24">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <SystemList
      :date-range="dateRange"
      :systems="systems"
      :max-height="maxHeight"
      @click:item="menu = false"
    />
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from '@vue/composition-api'

// Composables
import { useQuery } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@//use/systems'

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

    useQuery().sync({
      fromQuery(query) {
        if (query.system) {
          props.systems.change(query.system)
        }
      },
      toQuery() {
        if (props.systems.activeValue) {
          return { system: props.systems.activeValue }
        }
      },
    })

    watch(
      () => props.systems.list,
      () => {
        if (props.systems.activeItem) {
          return
        }

        if (props.systems.list.length) {
          props.systems.change(props.systems.list[0].system)
          return
        }
      },
      { immediate: true },
    )

    return { menu, attrs }
  },
})
</script>

<style lang="scss" scoped></style>
