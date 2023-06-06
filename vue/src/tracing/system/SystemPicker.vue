<template>
  <v-autocomplete
    v-autowidth="{ minWidth: 60 }"
    :value="value"
    :items="filteredSystems"
    item-value="system"
    item-text="system"
    :search-input.sync="searchInput"
    no-filter
    placeholder="system"
    prefix="system: "
    multiple
    clearable
    auto-select-first
    hide-details
    dense
    outlined
    background-color="light-blue lighten-5"
    class="v-select--fit"
    @click:clear="$emit('input', systems.length ? [systems[0].system] : [])"
  >
    <template #item="{ item, attrs }">
      <v-list-item v-bind="attrs" @click="toggleOne(item.system)">
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="value.includes(item.system)"
            @click.stop="toggle(item.system)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title :class="{ 'pl-4': item.indent }">{{ item.system }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action class="my-0">
          <v-list-item-action-text><XNum :value="item.groupCount" /></v-list-item-action-text>
        </v-list-item-action>
      </v-list-item>
    </template>
    <template #selection="{ index, item }">
      <div v-if="index === 2" class="v-select__selection">, {{ value.length - 2 }} more</div>
      <div v-else-if="index < 2" class="v-select__selection text-truncate">
        {{ comma(item, index) }}
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watchEffect, PropType } from 'vue'

// Composables
import { System } from '@/tracing/system/use-systems'

export default defineComponent({
  name: 'SystemPicker',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Array as PropType<string[]>,
      required: true,
    },
    systems: {
      type: Array as PropType<System[]>,
      required: true,
    },
    outlined: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const autocomplete = shallowRef()
    const searchInput = shallowRef('')

    const internalSystems = computed(() => {
      const systems = props.systems.slice()

      for (let system of props.value) {
        const index = systems.findIndex((item) => item.system === system)
        if (index === -1) {
          systems.push({
            system,
            count: 0,
            rate: 0,
            errorCount: 0,
            errorRate: 0,
            groupCount: 0,
          })
        }
      }

      return systems
    })

    const filteredSystems = computed(() => {
      if (!searchInput.value) {
        return internalSystems.value
      }
      return fuzzyFilter(internalSystems.value, searchInput.value, { key: 'system' })
    })

    watchEffect(
      () => {
        if (props.loading || props.value.length) {
          return
        }
        if (props.systems.length) {
          ctx.emit('input', props.systems[0].system)
        }
      },
      { flush: 'post' },
    )

    function comma(item: System, index: number): string {
      if (index > 0) {
        return ', ' + item.system
      }
      return item.system
    }

    function toggle(system: string) {
      let activeSystems = props.value.slice()
      const index = activeSystems.indexOf(system)

      if (index >= 0) {
        activeSystems.splice(index, 1)
        ctx.emit('input', activeSystems)
        return
      }

      if (system.endsWith(':all')) {
        ctx.emit('input', [system])
        return
      }

      if (activeSystems.length) {
        activeSystems = activeSystems.filter((system) => !system.endsWith(':all'))
      }
      activeSystems.push(system)

      ctx.emit('input', activeSystems)
    }

    function toggleOne(system: string) {
      const value = props.value.length === 1 && props.value.includes(system) ? [] : [system]
      ctx.emit('input', value)
    }

    return {
      autocomplete,
      searchInput,
      filteredSystems,
      comma,
      toggle,
      toggleOne,
    }
  },
})
</script>

<style lang="scss" scoped></style>
