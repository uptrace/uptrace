<template>
  <v-autocomplete
    ref="autocomplete"
    v-autowidth="{ minWidth: 60 }"
    :value="internalValue"
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
    background-color="bg--primary"
    :menu-props="{ maxHeight: 500 }"
    class="v-select--fit"
    @click:clear="$emit('input', [])"
  >
    <template #item="{ item, attrs }">
      <v-list-item v-bind="attrs" class="px-0">
        <v-menu :key="item.system" open-on-hover offset-x right :close-on-content-click="false">
          <template #activator="{ attrs, on }">
            <v-list-item v-bind="attrs" v-on="on" @click="toggleOne(item.system)">
              <v-list-item-action class="my-0 mr-4">
                <v-checkbox
                  :input-value="value.includes(item.system)"
                  @click.stop="toggle(item.system)"
                ></v-checkbox>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title :class="{ 'pl-4': item.indent }">
                  {{ item.system }} (<NumValue :value="item.groupCount" />)
                </v-list-item-title>
              </v-list-item-content>
              <v-list-item-icon>
                <v-icon v-if="item.children && item.children.length" dense>mdi-menu-right</v-icon>
              </v-list-item-icon>
            </v-list-item>
          </template>

          <v-card flat max-height="95vh">
            <v-list v-if="item.children && item.children.length" dense>
              <v-list-item
                v-for="child in item.children"
                :key="child.system"
                @click="toggleOne(child.system)"
              >
                <v-list-item-action class="my-0 mr-4">
                  <v-checkbox
                    :input-value="value.includes(child.system)"
                    @click.stop="toggle(child.system)"
                  ></v-checkbox>
                </v-list-item-action>
                <v-list-item-content>
                  <v-list-item-title>
                    {{ trimPrefix(child.system, item.system) }} (<NumValue
                      :value="child.groupCount"
                    />)
                  </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-menu>
      </v-list-item>
    </template>
    <template #selection="{ index, item }">
      <div v-if="index === 2" class="v-select__selection">, {{ value.length - 2 }} more</div>
      <div v-else-if="index < 2" class="v-select__selection text-truncate">
        {{ withComma(item, index) }}
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, watch, watchEffect, PropType } from 'vue'

// Composables
import { sortSystems, System } from '@/tracing/system/use-systems'

// Misc
import { truncateMiddle } from '@/util/string'
import { systemType, isGroupSystem, SystemName } from '@/models/otel'

const ALL_SYSTEMS = [
  SystemName.All,
  SystemName.SpansAll,
  SystemName.LogAll,
  SystemName.EventsAll,
] as string[]

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
    watch(
      () => autocomplete.value?.isMenuActive ?? false,
      (isMenuActive) => {
        if (!isMenuActive) {
          searchInput.value = ''
        }
      },
    )

    const internalValue = computed(() => {
      if (autocomplete.value?.isMenuActive) {
        return []
      }
      return props.value
    })

    const _tree = computed(() => {
      const systems = props.systems.slice()
      return buildSystemTree(systems)
    })
    const tree = shallowRef<SystemNode[]>([])
    watch(
      () => {
        if (autocomplete.value?.isMenuActive) {
          return undefined
        }

        const systems = _tree.value.slice()

        for (let system of props.value) {
          const foundIndex = systems.findIndex((item) => item.system === system)
          if (foundIndex >= 0) {
            continue
          }

          const type = systemType(system)
          const typeIndex = systems.findIndex((item) => item.system === type + ':all')

          const found = props.systems.find((item) => item.system === system)
          if (found) {
            systems.push({
              ...found,
              indent: typeIndex >= 0,
              children: [],
            })
            continue
          }

          systems.push({
            system,
            count: 0,
            rate: 0,
            errorCount: 0,
            errorRate: 0,
            groupCount: 0,
            indent: typeIndex >= 0,
            children: [],
          })
        }

        sortSystems(systems)

        return systems
      },
      (treeValue) => {
        if (treeValue) {
          tree.value = treeValue
        }
      },
      { immediate: true },
    )

    const filteredSystems = computed(() => {
      if (!searchInput.value) {
        return tree.value
      }
      return fuzzyFilter(props.systems, searchInput.value, { key: 'system' })
    })

    watchEffect(
      () => {
        if (props.value.length || props.loading) {
          return
        }
        if (props.systems.length) {
          ctx.emit('input', props.systems[0].system)
        }
      },
      { flush: 'post' },
    )

    function toggle(system: string) {
      let activeSystems = props.value.slice()

      const index = activeSystems.indexOf(system)
      if (index >= 0) {
        activeSystems.splice(index, 1)
        if (activeSystems.length) {
          ctx.emit('input', activeSystems)
          return
        }

        if (props.systems.length) {
          ctx.emit('input', [props.systems[0].system])
          return
        }

        ctx.emit('input', [])
        return
      }

      if (ALL_SYSTEMS.includes(system)) {
        ctx.emit('input', [system])
        return
      }

      activeSystems = activeSystems.filter((system) => !ALL_SYSTEMS.includes(system))

      if (isGroupSystem(system)) {
        const type = systemType(system)
        activeSystems = activeSystems.filter((system) => !system.startsWith(type + ':'))
      } else if (activeSystems.length) {
        const type = systemType(system)
        activeSystems = activeSystems.filter((system) => system !== type + ':all')
      }

      activeSystems.push(system)
      ctx.emit('input', activeSystems)
    }

    function toggleOne(system: string) {
      if (props.value.length === 1 && props.value.includes(system)) {
        if (props.systems.length) {
          ctx.emit('input', [props.systems[0].system])
          return
        }

        ctx.emit('input', [])
        return
      }

      ctx.emit('input', [system])
      return
    }

    return {
      autocomplete,
      searchInput,

      internalValue,
      filteredSystems,
      toggle,
      toggleOne,

      withComma,
      trimPrefix,
    }
  },
})

function withComma(item: System, index: number): string {
  const text = truncateMiddle(item.system, 20)
  if (index > 0) {
    return ', ' + text
  }
  return text
}

interface SystemNode extends System {
  indent?: boolean
  children?: System[]
}

function buildSystemTree(systems: System[]): SystemNode[] {
  const typeMap = new Map<string, SystemNode>()

  for (let sys of systems) {
    if (isGroupSystem(sys.system)) {
      continue
    }

    let typ = sys.system

    const i = typ.indexOf(':')
    if (i >= 0) {
      typ = typ.slice(0, i)
    }

    const typeSys = typeMap.get(typ)
    if (typeSys) {
      typeSys.count += sys.count
      typeSys.rate += sys.rate
      typeSys.errorCount += sys.errorCount
      typeSys.errorRate += sys.errorRate
      typeSys.groupCount += sys.groupCount
      typeSys.children!.push(sys)
      continue
    }

    typeMap.set(typ, {
      ...sys,
      system: typ + ':all',
      children: [sys],
    })
  }

  const nodes: SystemNode[] = []

  typeMap.forEach((system) => {
    if (system.children!.length === 1) {
      nodes.push({
        ...system.children![0],
      })
    } else {
      nodes.push(system)
    }
  })

  if (nodes.length === 1 && nodes[0].children) {
    return nodes[0].children
  }
  return nodes
}

function trimPrefix(str: string, prefix: string): string {
  if (prefix.endsWith(':all')) {
    prefix = prefix.slice(0, -3)
  }
  if (str.startsWith(prefix)) {
    return str.slice(prefix.length)
  }
  return str
}
</script>

<style lang="scss" scoped></style>
