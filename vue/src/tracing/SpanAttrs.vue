<template>
  <div>
    <v-row align="center" dense>
      <v-col cols="3" md="2">
        <v-text-field
          v-model="searchInput"
          label="Filter keys"
          clearable
          dense
          outlined
          hide-details
          class="pt-0"
        />
      </v-col>

      <v-col cols="9" md="10">
        <v-slide-group v-model="activePrefix" center-active show-arrows>
          <v-slide-item
            v-for="(item, i) in prefixes"
            :key="item.prefix"
            v-slot="{ active, toggle }"
            :value="item"
          >
            <v-btn
              :input-value="active"
              active-class="light-blue white--text"
              small
              depressed
              rounded
              class="text-transform-none"
              :class="{ 'ml-1': i > 0 }"
              @click="toggle"
            >
              {{ item.prefix }}
            </v-btn>
          </v-slide-item>
        </v-slide-group>
      </v-col>
    </v-row>

    <v-row dense align="center">
      <v-col cols="auto">
        <v-tabs v-model="activeTab" class="align-center">
          <v-tab href="#table">Table</v-tab>
          <v-tab href="#json">JSON</v-tab>

          <v-tab v-if="dbStmtPretty" href="#dbStmtPretty">SQL:pretty</v-tab>
          <v-tab v-if="dbStmt" href="#dbStmt">
            {{ dbStmtPretty ? 'SQL:original' : AttrKey.dbStatement }}
          </v-tab>
          <v-tab v-if="exceptionStacktrace" href="#exceptionStacktrace">Stacktrace</v-tab>

          <v-tab
            v-for="(attrValue, attrKey) in largeAttrs"
            :key="attrKey"
            :href="`#attr-${attrKey}`"
            >{{ attrKey }}</v-tab
          >
        </v-tabs>
      </v-col>
    </v-row>

    <v-row dense>
      <v-col>
        <v-tabs-items v-model="activeTab">
          <v-tab-item value="table">
            <SpanAttrsTable
              :date-range="dateRange"
              :attrs="attrs"
              :attr-keys="attrKeys"
              :system="system"
              :group-id="groupId"
            />
          </v-tab-item>
          <v-tab-item value="json">
            <PrismCode :code="prettyPrint(filteredAttrs)" language="json" />
          </v-tab-item>

          <v-tab-item value="dbStmtPretty">
            <PrismCode :code="dbStmtPretty" language="sql" />
          </v-tab-item>
          <v-tab-item value="dbStmt">
            <PrismCode v-if="dbStmtJson" :code="dbStmtJson" language="json" />
            <PrismCode v-else :code="dbStmt" language="sql" />
          </v-tab-item>
          <v-tab-item value="exceptionStacktrace">
            <PrismCode :code="exceptionStacktrace" />
          </v-tab-item>

          <v-tab-item
            v-for="(attrValue, attrKey) in largeAttrs"
            :key="attrKey"
            :value="`attr-${attrKey}`"
          >
            <PrismCode :code="attrValue" />
          </v-tab-item>
        </v-tabs-items>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { format, supportedDialects } from 'sql-formatter'
import { pick } from 'lodash-es'
import { defineComponent, shallowRef, computed, PropType } from 'vue'
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useProject } from '@/org/use-projects'

// Components
import SpanAttrsTable from '@/tracing/SpanAttrsTable.vue'

// Misc
import { AttrMap } from '@/models/span'
import { AttrKey } from '@/models/otel'
import { buildPrefixes, Prefix } from '@/models/key-prefixes'
import { parseJson, prettyPrint } from '@/util/json'

const specialKeys = [AttrKey.dbStatement, AttrKey.exceptionStacktrace] as string[]

const LARGE_ATTR_THRESHOLD = 500

export default defineComponent({
  name: 'SpanAttrs',
  components: { SpanAttrsTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    attrs: {
      type: Object as PropType<AttrMap>,
      required: true,
    },
    system: {
      type: String,
      default: undefined,
    },
    groupId: {
      type: String,
      default: undefined,
    },
  },

  setup(props) {
    const activeTab = shallowRef()
    const searchInput = shallowRef('')

    const activePrefix = shallowRef<Prefix>()
    const prefixes = computed(() => {
      const keys = []
      for (let key in props.attrs) {
        if (!isBlacklistedKey(key)) {
          keys.push(key)
        }
      }
      return buildPrefixes(keys)
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        group_id: props.groupId,
      }
    })

    const dbStmt = computed((): string => {
      return props.attrs[AttrKey.dbStatement] ?? ''
    })

    const sqlLanguage = computed(() => {
      const system = props.attrs[AttrKey.dbSystem]
      if (!system) {
        return 'sql'
      }

      if (supportedDialects.indexOf(system) >= 0) {
        return system
      }

      return 'sql'
    })

    const dbStmtPretty = computed((): string => {
      try {
        return format(dbStmt.value, {
          language: sqlLanguage.value,
        })
      } catch (err) {
        return ''
      }
    })

    const dbStmtJson = computed((): any => {
      const obj = parseJson(props.attrs[AttrKey.dbStatement])
      if (!obj) {
        return undefined
      }
      return prettyPrint(obj)
    })

    const exceptionStacktrace = computed((): string => {
      return props.attrs[AttrKey.exceptionStacktrace] ?? ''
    })

    const project = useProject()

    const largeAttrs = computed((): Record<string, string> => {
      const attrs: Record<string, string> = {}

      for (let key in props.attrs) {
        if (isInternalKey(key) || specialKeys.includes(key)) {
          continue
        }

        const value = props.attrs[key]

        const json = parseJson(value)
        if (json) {
          const pretty = prettyPrint(json)
          if (project.largeAttrs.includes(key) || pretty.length >= LARGE_ATTR_THRESHOLD) {
            attrs[key] = prettyPrint(json)
            continue
          }
        }

        if (project.largeAttrs.includes(key)) {
          attrs[key] = value
        }

        switch (typeof value) {
          case 'string':
            if (value.length >= LARGE_ATTR_THRESHOLD) {
              attrs[key] = value
            }
        }
      }

      return attrs
    })

    const attrKeys = computed((): string[] => {
      let keys = Object.keys(props.attrs)
      keys = keys.filter((key) => !isBlacklistedKey(key))

      if (activePrefix.value) {
        keys = activePrefix.value.keys
      }

      if (searchInput.value) {
        keys = fuzzyFilter(keys, searchInput.value)
      }

      keys.sort()

      return keys
    })

    const filteredAttrs = computed(() => {
      return pick(props.attrs, attrKeys.value)
    })

    function isBlacklistedKey(key: string): boolean {
      return isInternalKey(key) || key in largeAttrs.value || specialKeys.includes(key)
    }

    return {
      AttrKey,

      activeTab,
      searchInput,
      activePrefix,
      prefixes,

      axiosParams,

      dbStmt,
      dbStmtPretty,
      dbStmtJson,
      exceptionStacktrace,

      largeAttrs,
      attrKeys,
      filteredAttrs,
      prettyPrint,
    }
  },
})

function isInternalKey(key: string): boolean {
  return key.startsWith('_')
}
</script>

<style lang="scss" scoped></style>
