<template>
  <v-btn-toggle :value="activeGroupSystem" group color="blue accent-3">
    <v-btn
      v-for="group in groups"
      :key="group.system"
      :value="group.system"
      @click="$router.push(routeFor(group.system)).catch(() => {})"
    >
      {{ group.name }} ({{ group.count }})
    </v-btn>
  </v-btn-toggle>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { createUqlEditor, useQueryStore } from '@/use/uql'
import { addAllSystem, System } from '@/tracing/system/use-systems'

// Utilities
import { isSpanSystem, isEventSystem, isLogSystem, SystemName, AttrKey } from '@/models/otel'

interface Group {
  name: string
  system: SystemName
  count: number
}

export default defineComponent({
  name: 'SystemGroupPicker',

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    systems: {
      type: Array as PropType<System[]>,
      required: true,
    },
    value: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const spanSystems = computed(() => {
      const systems = props.systems.filter((item) => isSpanSystem(item.system))
      addAllSystem(systems, SystemName.SpansAll)
      return systems
    })

    const logSystems = computed(() => {
      const systems = props.systems.filter((item) => isLogSystem(item.system))
      addAllSystem(systems, SystemName.LogAll)
      return systems
    })

    const eventSystems = computed(() => {
      const systems = props.systems.filter(
        (item) => isEventSystem(item.system) && !isLogSystem(item.system),
      )
      addAllSystem(systems, SystemName.EventsAll)
      return systems
    })

    const groups = computed(() => {
      const groups: Group[] = []

      if (spanSystems.value.length > 1) {
        groups.push({
          name: 'Spans',
          system: SystemName.SpansAll,
          count: countGroups(spanSystems.value),
        })
      }

      if (logSystems.value.length > 1) {
        groups.push({
          name: 'Logs',
          system: SystemName.LogAll,
          count: countGroups(logSystems.value),
        })
      }

      if (eventSystems.value.length > 1) {
        groups.push({
          name: 'Events',
          system: SystemName.EventsAll,
          count: countGroups(eventSystems.value),
        })
      }

      return groups
    })

    const activeGroupSystem = computed(() => {
      if (!props.value.length) {
        if (groups.value.length) {
          return groups.value[0].system
        }
        return undefined
      }

      const system = props.value[0]
      if (isLogSystem(system)) {
        return SystemName.LogAll
      }
      if (isEventSystem(system)) {
        return SystemName.EventsAll
      }
      return SystemName.SpansAll
    })

    const systemItems = computed(() => {
      if (props.loading) {
        return []
      }

      switch (activeGroupSystem.value) {
        case SystemName.SpansAll:
          return spanSystems.value
        case SystemName.EventsAll:
          return eventSystems.value
        case SystemName.LogAll:
          return logSystems.value
        default:
          return []
      }
    })

    watch(
      systemItems,
      (systemItems) => {
        ctx.emit('update:systems', systemItems)
      },
      { flush: 'sync' },
    )

    const { where } = useQueryStore()
    function routeFor(system: string) {
      return {
        name: 'SpanGroupList',
        query: {
          system,
          query: createUqlEditor()
            .exploreAttr(AttrKey.spanGroupId, isEventSystem(system))
            .add(where.value)
            .toString(),
        },
      }
    }

    return { groups, activeGroupSystem, routeFor }
  },
})

function countGroups(systems: System[]) {
  let sum = 0
  for (let system of systems) {
    if (!system.isGroup) {
      sum += system.groupCount
    }
  }
  return sum
}
</script>

<style lang="scss" scoped></style>
