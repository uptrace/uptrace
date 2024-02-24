<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">Search</v-btn>
    </template>
    <v-form @submit.prevent="addFilter">
      <v-card width="550">
        <v-card-text class="pa-6">
          <v-row>
            <v-col>
              <v-btn-toggle v-model="activeItem" dense tile color="primary">
                <v-btn v-for="item in items" :key="item.value" :value="item">
                  {{ item.value }}
                </v-btn>
              </v-btn-toggle>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <template v-if="activeItem">
                <v-chip v-for="attr in activeItem.attrs" :key="attr" class="mr-1">{{
                  attr
                }}</v-chip>
              </template>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                v-model="searchInput"
                label="Contains substr1|substr2|substr3"
                hint='Case-insensitive options separated with "|"'
                persistent-hint
                filled
                dense
                autofocus
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn type="submit" class="primary" :disabled="!isValid">Filter</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watchEffect, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'

// Misc
import { AttrKey, SystemName } from '@/models/otel'
import { quote, escapeRe } from '@/util/string'

export default defineComponent({
  name: 'SearchFilterMenu',

  props: {
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const { router, route } = useRouter()
    const menu = shallowRef(false)
    const activeItem = shallowRef()
    const searchInput = shallowRef('')

    const items = computed(() => {
      return [
        {
          value: 'any',
          attrs: [AttrKey.displayName],
          system: SystemName.SpansAll,
        },
        {
          value: 'spans',
          attrs: [AttrKey.spanName],
          system: SystemName.SpansAll,
        },
        {
          value: 'events',
          attrs: [AttrKey.spanEventName],
          system: SystemName.EventsAll,
        },
        {
          value: 'http',
          attrs: [AttrKey.httpMethod, AttrKey.httpRoute, AttrKey.httpTarget],
          system: SystemName.HttpAll,
        },
        {
          value: 'logs',
          attrs: [AttrKey.logSeverity, AttrKey.logMessage],
          system: SystemName.LogAll,
        },
        {
          value: 'exceptions',
          attrs: [AttrKey.exceptionType, AttrKey.exceptionMessage],
          system: SystemName.LogAll,
        },
        {
          value: 'funcs',
          attrs: [AttrKey.codeFunction, AttrKey.codeFilepath],
          system: SystemName.SpansAll,
        },
        {
          value: 'db',
          attrs: [AttrKey.dbOperation, AttrKey.dbSqlTable, AttrKey.dbStatement],
          system: SystemName.DbAll,
        },
      ]
    })

    const isValid = computed(() => {
      return activeItem.value && searchInput.value
    })

    watchEffect(() => {
      if (!activeItem.value && items.value.length) {
        activeItem.value = items.value[0]
      }
    })

    function addFilter() {
      if (!isValid.value || !activeItem.value) {
        menu.value = false
        return
      }

      let system: any = activeItem.value.system
      if (!system) {
        system = props.systems.activeSystems
      }

      const attrs = activeItem.value.attrs
      const key = attrs.length > 1 ? `{${attrs.join(',')}}` : attrs[0]
      const quotedValue = quote(searchInput.value)

      const editor = props.uql.createEditor()
      editor.replaceOrPush(
        new RegExp(`^where\\s+${escapeRe(key)}\\s+contains\\s+.+`, 'i'),
        `where ${key} contains ${quotedValue}`,
      )
      const query = editor.toString()

      router.push({
        name: 'SpanGroupList',
        query: {
          ...route.value.query,
          system,
          query,
        },
      })

      menu.value = false
    }

    return {
      AttrKey,
      menu,

      activeItem,
      items,
      searchInput,
      isValid,

      addFilter,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-btn-toggle ::v-deep .v-btn {
  text-transform: none;
}
</style>
