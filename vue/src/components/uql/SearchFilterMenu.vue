<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>Search</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-form @submit.prevent="addFilter">
      <v-card width="550">
        <v-card-text class="pa-6">
          <v-row>
            <v-col>
              <v-btn-toggle v-model="attrSet.active" dense tile color="primary">
                <v-btn v-for="item in attrSet.items" :key="item.value" :value="item.value">
                  {{ item.value }}
                </v-btn>
              </v-btn-toggle>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                v-model="attrKeys"
                :items="attrKeyItems"
                :rules="rules"
                hide-details="auto"
                chips
                small-chips
                multiple
                filled
              >
              </v-autocomplete>
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
import { isEqual } from 'lodash'
import {
  defineComponent,
  proxyRefs,
  shallowRef,
  computed,
  Ref,
  PropType,
} from '@vue/composition-api'

// Composables
import { UseUql } from '@/use/uql'

// Utilities
import { xkey } from '@/models/otelattr'
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
    const menu = shallowRef(false)

    const rules = [
      (v: string[]) => {
        if (!v || !v.length) {
          return 'Please select at least one attribute'
        }
        if (v.length > 4) {
          return 'You can search at most over 4 attributes'
        }
        return true
      },
    ]

    const isValid = computed(() => {
      return attrKeys.value && attrKeys.value.length && attrValue.value
    })

    const attrValue = shallowRef('')
    const attrKeys = shallowRef<string[]>([xkey.spanName, xkey.spanEventName])

    const attrKeyItems = [
      xkey.spanName,
      xkey.spanEventName,
      xkey.exceptionType,
      xkey.exceptionMessage,
      xkey.logSeverity,
      xkey.logMessage,
      xkey.codeFunction,
      xkey.codeFilepath,
      xkey.dbOperation,
      xkey.dbSqlTables,
      xkey.dbStatement,
    ]

    const isEventAttrKey = computed(() => {
      return attrKeys.value.some((attrKey) => {
        const candidates = [
          xkey.exceptionType,
          xkey.exceptionMessage,
          xkey.logSeverity,
          xkey.logMessage,
        ]
        return candidates.indexOf(attrKey as xkey) >= 0
      })
    })

    function addFilter() {
      if (!isValid.value) {
        menu.value = false
        return
      }

      const key = attrKeys.value.join(',')
      const quotedValue = quote(attrValue.value)

      const editor = props.uql.createEditor()
      editor.add(`where {${key}} contains ${quotedValue}`)
      if (isEventAttrKey.value) {
        editor.add(`where ${xkey.spanIsEvent}`)
      }
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      xkey,
      menu,

      attrKeys,
      attrSet: useAttrSet(attrKeys),
      attrKeyItems,
      attrValue,
      rules,
      isValid,

      addFilter,
    }
  },
})

function useAttrSet(attrKeys: Ref<string[]>) {
  const items = computed(() => {
    return [
      {
        value: 'span',
        attrKeys: [xkey.spanName, xkey.spanEventName],
      },
      {
        value: 'log or error',
        attrKeys: [xkey.exceptionType, xkey.exceptionMessage, xkey.logSeverity, xkey.logMessage],
      },
      {
        value: 'code',
        attrKeys: [xkey.codeFunction, xkey.codeFilepath],
      },
      {
        value: 'db',
        attrKeys: [xkey.dbOperation, xkey.dbSqlTables, xkey.dbStatement],
      },
    ]
  })

  const active = computed({
    get() {
      for (let attrSet of items.value) {
        if (isEqual(attrSet.attrKeys, attrKeys.value)) {
          return attrSet.value
        }
      }
      return ''
    },
    set(value: string) {
      const item = items.value.find((item) => item.value === value)
      attrKeys.value = item?.attrKeys ?? []
    },
  })

  return proxyRefs({ items, active })
}
</script>

<style lang="scss" scoped>
.v-btn-toggle :deep(.v-btn) {
  text-transform: none;
}
</style>
