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

    <v-row dense>
      <v-col>
        <v-tabs>
          <v-tab>Table</v-tab>
          <v-tab>JSON</v-tab>

          <v-tab-item>
            <SpanAttrsTable
              :date-range="dateRange"
              :attrs="attrs"
              :attr-keys="attrKeys"
              :system="system"
              :group-id="groupId"
            />
          </v-tab-item>
          <v-tab-item>
            <PrismCode :code="prettyPrint(nestedAttrs)" language="json" />
          </v-tab-item>
        </v-tabs>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import SpanAttrsTable from '@/tracing/SpanAttrsTable.vue'

// Utitlies
import { AttrMap } from '@/models/span'
import { AttrKey, isEventSystem } from '@/models/otel'
import { buildPrefixes, Prefix } from '@/models/key-prefixes'
import { parseJson, prettyPrint } from '@/util/json'

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
      default: "",
    },
    groupId: {
      type: String,
      default: "",
    },
  },

  setup(props) {
    const searchInput = shallowRef('')

    const activePrefix = shallowRef<Prefix>()
    const prefixes = computed(() => {
      const keys = []
      for (let key in props.attrs) {
        keys.push(key)
      }
      return buildPrefixes(keys)
    })

    const isEvent = computed((): boolean => {
      return isEventSystem(props.system)
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        group_id: props.groupId,
      }
    })

    const attrKeys = computed((): string[] => {
      let keys = Object.keys(props.attrs)

      if (activePrefix.value) {
        keys = activePrefix.value.keys
      }

      if (searchInput.value) {
        keys = fuzzyFilter(keys, searchInput.value)
      }

      keys.sort()

      return keys
    })

    const nestedAttrs = computed(() => {
      return nestAttrs(props.attrs, attrKeys.value)
    })

    return {
      AttrKey,

      searchInput,
      activePrefix,
      prefixes,

      axiosParams,
      isEvent,

      attrKeys,
      nestedAttrs,
      prettyPrint,
    }
  },
})

function nestAttrs(src: Record<string, any>, keys: string[]) {
  const dest: Record<string, any> = {}
  for (let key of keys) {
    if (key.indexOf('.') === -1) {
      dest[key] = src[key]
    } else {
      setValue(dest, key.split('.'), src[key])
    }
  }
  return dest
}

function setValue(dest: Record<string, any>, path: string[], value: any) {
  for (let key of path.slice(0, -1)) {
    if (!(key in dest)) {
      const v = {}
      dest[key] = v
      dest = v
      continue
    }

    let v = dest[key]
    if (v === null || typeof v !== 'object' || Array.isArray(v)) {
      v = { _value: v }
      dest[key] = v
    }
    dest = v
  }

  if (typeof value === 'string') {
    const data = parseJson(value)
    if (data) {
      value = data
    }
  }

  const lastKey = path[path.length - 1]
  dest[lastKey] = value
}
</script>

<style lang="scss" scoped></style>
