<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn :disabled="disabled" text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>Agg</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-form ref="form" v-model="isValid" @submit.prevent="addFilter">
      <v-card width="400px">
        <v-card-text class="py-6 text-body-2">
          <v-row>
            <v-col class="no-transform space-around">
              <UqlChip
                v-for="col in aggColumns"
                :key="col.name"
                :uql="uql"
                :column="col.name"
                :tooltip="col.tooltip"
                @click="menu = false"
              />
            </v-col>
          </v-row>

          <div class="mt-3 mb-4 d-flex align-center">
            <v-divider />
            <div class="mx-2 grey--text text--lighten-1">or</div>
            <v-divider />
          </div>

          <v-row class="mb-n1">
            <v-col> Select a function and then a column: <strong>func(column)</strong>. </v-col>
          </v-row>
          <v-row dense class="mb-n6">
            <v-col :cols="5">
              <v-autocomplete
                v-model="func"
                :rules="rules.func"
                label="Function"
                :items="funcItems"
                return-object
                outlined
                dense
                class="fit"
              ></v-autocomplete>
            </v-col>
            <v-col :cols="7">
              <SimpleSuggestions
                v-model="column"
                :loading="suggestions.loading"
                :suggestions="suggestions"
                label="Column"
                :rules="rules.column"
                :disabled="!func || func.isColumn"
                dense
                class="fit"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn type="submit" :disabled="!isValid" class="primary">Add column</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { useSuggestions, Suggestion } from '@/use/suggestions'
import { UseUql } from '@/use/uql'

// Components
import SimpleSuggestions from '@/components/SimpleSuggestions.vue'
import UqlChip from '@/components/UqlChip.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { requiredRule } from '@/util/validation'

interface FuncItem {
  text: string
  value: string
}

const aggFuncs = [
  'any',
  'top3',
  'top10',
  'avg',
  'uniq',
  'p50',
  'p75',
  'p90',
  'p95',
  'p99',
  'min',
  'max',
  'sum',
]
const aggColumns = [
  { name: xkey.spanCount, tooltip: 'Number of spans in a group' },
  { name: xkey.spanCountPerMin, tooltip: 'Number of spans per minute in a group' },
  { name: xkey.spanErrorCount, tooltip: 'Number of spans with span.status_code = "error"' },
  { name: xkey.spanErrorPct, tooltip: 'Percent of spans with span.status_code = "error"' },
  { name: `p75(${xkey.spanDuration})`, tooltip: '75th percentile of span.duration' },
  { name: `max(${xkey.spanDuration})`, tooltip: 'Max span.duration' },
  { name: `uniq(${xkey.enduserId})`, tooltip: 'Number of distinct enduser.id' },
]

export default defineComponent({
  name: 'AggFilterMenu',
  components: { SimpleSuggestions, UqlChip },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const menu = shallowRef(false)
    const func = shallowRef<FuncItem>()
    const column = shallowRef<Suggestion>()

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = {
      func: [requiredRule],
      column: [isColumnValid],
    }

    const suggestions = useSuggestions(
      () => {
        if (!menu.value) {
          return null
        }

        const { projectId } = route.value.params
        return {
          url: `/api/tracing/${projectId}/suggestions/attributes`,
          params: {
            ...props.axiosParams,
            func: func.value?.value,
          },
        }
      },
      { suggestSearchInput: true },
    )

    const funcItems = computed((): FuncItem[] => {
      const items: FuncItem[] = []

      for (let func of aggFuncs) {
        items.push({ value: func, text: func + '(...)' })
      }

      return items
    })

    function addFilter() {
      if (!func.value || !column.value) {
        return
      }

      const query = `${func.value.value}(${column.value.text})`
      aggBy(query)

      func.value = undefined
      column.value = undefined
      form.value.resetValidation()
    }

    function aggBy(column: string) {
      const editor = props.uql.createEditor()
      editor.add(column)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    function isColumnValid(s: any) {
      const isValid = Boolean(func.value && s)
      return isValid || 'Column is required'
    }

    return {
      menu,

      form,
      isValid,
      rules,
      suggestions,

      aggColumns,
      func,
      funcItems,
      column,

      addFilter,
      aggBy,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-select.fit {
  min-width: min-content !important;
}

.v-select.fit .v-select__selection--comma {
  text-overflow: unset;
}

.no-transform ::v-deep .v-btn {
  padding: 0 12px !important;
  text-transform: none;
}

.space-around ::v-deep .v-chip {
  margin: 4px;
}
</style>
