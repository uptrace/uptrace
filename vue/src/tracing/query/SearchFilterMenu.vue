<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on"> Search </v-btn>
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
import { isEqual } from 'lodash-es'
import { defineComponent, proxyRefs, shallowRef, computed, Ref, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseUql } from '@/use/uql'

// Utilities
import { AttrKey, SystemName } from '@/models/otelattr'
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
    const attrKeys = shallowRef<string[]>([AttrKey.spanName, AttrKey.spanEventName])

    const attrKeyItems = [
      AttrKey.spanName,
      AttrKey.spanEventName,
      AttrKey.exceptionType,
      AttrKey.exceptionMessage,
      AttrKey.logSeverity,
      AttrKey.logMessage,
      AttrKey.codeFunction,
      AttrKey.codeFilepath,
      AttrKey.dbOperation,
      AttrKey.dbSqlTables,
      AttrKey.dbStatement,
    ]

    function addFilter() {
      if (!isValid.value) {
        menu.value = false
        return
      }

      const key = attrKeys.value.join(',')
      const quotedValue = quote(attrValue.value)

      const editor = props.uql.createEditor()
      editor.add(`where {${key}} contains ${quotedValue}`)
      router.push({
        query: {
          ...route.value.query,
          system: SystemName.all, // TODO: pick a better system
          query: editor.toString(),
        },
      })

      menu.value = false
    }

    return {
      AttrKey,
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
        attrKeys: [AttrKey.spanName, AttrKey.spanEventName],
      },
      {
        value: 'log',
        attrKeys: [AttrKey.logSeverity, AttrKey.logMessage],
      },
      {
        value: 'exception',
        attrKeys: [AttrKey.exceptionType, AttrKey.exceptionMessage],
      },
      {
        value: 'code',
        attrKeys: [AttrKey.codeFunction, AttrKey.codeFilepath],
      },
      {
        value: 'db',
        attrKeys: [AttrKey.dbOperation, AttrKey.dbSqlTables, AttrKey.dbStatement],
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
.v-btn-toggle ::v-deep .v-btn {
  text-transform: none;
}
</style>
