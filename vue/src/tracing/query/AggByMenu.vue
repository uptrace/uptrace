<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on"> Agg </v-btn>
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
            <v-col>Select a function and then a column: <strong>func(column)</strong>.</v-col>
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
              <Combobox
                v-model="column"
                :data-source="suggestions"
                label="Column"
                :rules="rules.column"
                :disabled="!func"
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
import { useDataSource, Item } from '@/use/datasource'
import { UseUql } from '@/use/uql'

// Components
import Combobox from '@/components/Combobox.vue'
import UqlChip from '@/components/UqlChip.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { requiredRule } from '@/util/validation'

const AGG_FUNCS = [
  { name: 'sum' },
  { name: 'avg' },
  { name: 'min' },
  { name: 'max' },
  { name: 'uniq' },

  { name: 'any' },
  { name: 'anyLast' },
  { name: 'top3' },
  { name: 'top10' },

  { name: 'p50' },
  { name: 'p75' },
  { name: 'p90' },
  { name: 'p95' },
  { name: 'p99' },
]
interface FuncItem {
  text: string
  value: string
}

const aggColumns = [
  { name: AttrKey.spanCount, tooltip: 'Number of spans in a group' },
  { name: AttrKey.spanCountPerMin, tooltip: 'Number of spans per minute in a group' },
  { name: AttrKey.spanErrorCount, tooltip: 'Number of spans with span.status_code = "error"' },
  { name: AttrKey.spanErrorPct, tooltip: 'Percent of spans with span.status_code = "error"' },
  { name: `p75(${AttrKey.spanDuration})`, tooltip: '75th percentile of span.duration' },
  { name: `max(${AttrKey.spanDuration})`, tooltip: 'Max span.duration' },
  { name: `uniq(${AttrKey.enduserId})`, tooltip: 'Number of distinct enduser.id' },
]

export default defineComponent({
  name: 'AggFilterMenu',
  components: { Combobox, UqlChip },

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
    const column = shallowRef<Item>()

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = {
      func: [requiredRule],
      column: [isColumnValid],
    }

    const suggestions = useDataSource(
      () => {
        if (!menu.value) {
          return null
        }

        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/attr-keys`,
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
      for (let func of AGG_FUNCS) {
        items.push({ value: func.name, text: func.name + '(...)' })
      }
      return items
    })

    function addFilter() {
      if (!func.value || !column.value) {
        return
      }

      const query = `${func.value.value}(${column.value.value})`
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
