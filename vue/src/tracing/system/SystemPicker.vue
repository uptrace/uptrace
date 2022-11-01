<template>
  <v-autocomplete
    ref="autocomplete"
    v-autowidth="{ minWidth: '60px' }"
    :value="value"
    :items="filteredItems"
    item-value="system"
    item-text="system"
    :search-input.sync="searchInput"
    no-filter
    placeholder="system"
    multiple
    clearable
    auto-select-first
    hide-details
    dense
    outlined
    background-color="light-blue lighten-5"
    class="v-select--fit"
    @click:clear="$emit('input', [allSystem])"
  >
    <template #item="{ item, attrs }">
      <v-list-item
        v-bind="attrs"
        @click="
          $emit('input', [item.system])
          autocomplete.blur()
        "
      >
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="value.indexOf(item.system) >= 0"
            dense
            @click.stop="toggleSystem(item.system)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>{{ item.system }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action class="my-0">
          <v-list-item-action-text><XNum :value="item.count" /></v-list-item-action-text>
        </v-list-item-action>
      </v-list-item>
    </template>
    <template #selection="{ index, item }">
      <div v-if="index === 3" class="v-select__selection">, {{ value.length - 3 }} more</div>
      <div v-else-if="index < 3" class="v-select__selection text-truncate">
        {{ comma(item, index) }}
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watchEffect, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { System } from '@/tracing/system/use-systems'

// Utilities
import { splitTypeSystem } from '@/models/otel'

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

  setup(props, ctx) {
    const route = useRoute()
    const autocomplete = shallowRef()
    const searchInput = shallowRef('')

    const internalItems = computed(() => {
      const items = props.items.slice()

      const allSystem = {
        projectId: route.value.params.projectId,
        system: props.allSystem,
        count: 0,
        rate: 0,
        errorCount: 0,
        errorPct: 0,
      }
      for (let item of items) {
        if (!item.system.endsWith(':all')) {
          allSystem.count += item.count
          allSystem.rate += item.rate
          allSystem.errorCount += item.errorCount
        }
      }
      allSystem.errorPct = allSystem.errorCount / allSystem.count
      items.unshift(allSystem)

      for (let system of props.value) {
        const index = items.findIndex((item) => item.system === system)
        if (index === -1) {
          items.push({
            projectId: route.value.params.projectId,
            system,
            count: 0,
            rate: 0,
            errorCount: 0,
            errorPct: 0,
          })
        }
      }

      return items
    })

    const filteredItems = computed(() => {
      if (!searchInput.value) {
        return internalItems.value
      }
      return fuzzyFilter(internalItems.value, searchInput.value, { key: 'system' })
    })

    watchEffect(
      () => {
        if (props.value.length) {
          return
        }
        if (internalItems.value.length) {
          ctx.emit('input', internalItems.value[0].system)
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

    function toggleSystem(system: string) {
      let activeSystems = props.value.slice() as string[]
      const index = activeSystems.indexOf(system)

      if (index >= 0) {
        activeSystems.splice(index, 1)
        ctx.emit('input', activeSystems)
        return
      }

      if (system === props.allSystem) {
        ctx.emit('input', [props.allSystem])
        return
      }

      if (activeSystems.length) {
        const index = activeSystems.indexOf(props.allSystem)
        if (index >= 0) {
          activeSystems.splice(index, 1)
        }
      }

      if (system.endsWith(':all')) {
        activeSystems = tryRemoveChildren(activeSystems, system)
      } else {
        activeSystems = tryRemoveAllSystem(activeSystems, system)
      }

      activeSystems.push(system)
      ctx.emit('input', activeSystems)
    }

    function tryRemoveChildren(systems: string[], needle: string) {
      const prefix = splitTypeSystem(needle)[0] + ':'
      return systems.filter((system) => system === needle || !system.startsWith(prefix))
    }

    function tryRemoveAllSystem(systems: string[], needle: string) {
      needle = splitTypeSystem(needle)[0] + ':all'
      const index = systems.indexOf(needle)
      if (index >= 0) {
        systems.splice(index, 1)
      }
      return systems
    }

    return {
      autocomplete,
      searchInput,
      filteredItems,
      comma,
      toggleSystem,
    }
  },
})
</script>

<style lang="scss" scoped></style>
