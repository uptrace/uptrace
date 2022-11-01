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
                v-model="attrValue"
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
import { UseUql } from '@/use/uql'

// Utilities
import { isEventSystem, AttrKey, SystemName } from '@/models/otel'
import { quote } from '@/util/string'

export default defineComponent({
  name: 'SearchFilterMenu',

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const { router, route } = useRouter()
    const menu = shallowRef(false)
    const activeItem = shallowRef()
    const attrValue = shallowRef('')

    const items = computed(() => {
      return [
        {
          value: 'spans',
          attrs: [AttrKey.spanName],
          system: SystemName.spansAll,
        },
        {
          value: 'events',
          attrs: [AttrKey.spanEventName],
          system: SystemName.eventsAll,
        },
        {
          value: 'http',
          attrs: [AttrKey.httpMethod, AttrKey.httpRoute, AttrKey.httpTarget],
          system: SystemName.httpAll,
        },
        {
          value: 'logs',
          attrs: [AttrKey.logSeverity, AttrKey.logMessage],
          system: SystemName.logAll,
        },
        {
          value: 'exceptions',
          attrs: [AttrKey.exceptionType, AttrKey.exceptionMessage],
          system: SystemName.exceptions,
        },
        {
          value: 'code',
          attrs: [AttrKey.codeFunction, AttrKey.codeFilepath],
          system: SystemName.spansAll,
        },
        {
          value: 'db',
          attrs: [AttrKey.dbOperation, AttrKey.dbSqlTables, AttrKey.dbStatement],
          system: SystemName.dbAll,
        },
      ]
    })

    const isValid = computed(() => {
      return activeItem.value && attrValue.value
    })

    watchEffect(() => {
      if (!activeItem.value && items.value.length) {
        activeItem.value = items.value[0]
      }
    })

    function addFilter() {
      if (!isValid.value) {
        menu.value = false
        return
      }

      const { attrs, system } = activeItem.value
      const key = attrs.length > 1 ? `{${attrs.join(',')}}` : attrs[0]
      const quotedValue = quote(attrValue.value)

      const editor = props.uql.createEditor()
      editor.add(`where ${key} contains ${quotedValue}`)
      const query = editor.toString()

      router.push({
        name: isEventSystem(system) ? 'EventGroupList' : 'SpanGroupList',
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
      attrValue,
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
