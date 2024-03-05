<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-list-item v-bind="attrs" v-on="on">
        <v-list-item-title>Where</v-list-item-title>
      </v-list-item>
    </template>
    <v-card class="pa-3">
      <v-row>
        <v-col cols="auto">
          <v-form ref="formRef" v-model="form.isValid" class="pa-3" style="width: 400px">
            <v-row>
              <v-col class="space-around">
                <UqlChip
                  :uql="uql"
                  column="span.event_count"
                  op=">"
                  value="0"
                  tooltip="Filter spans with events"
                  @click="menu = false"
                />
                <UqlChip
                  :uql="uql"
                  column="span.event_log_count"
                  op=">"
                  value="0"
                  tooltip="Filter spans with logs"
                  @click="menu = false"
                />
                <UqlChip
                  :uql="uql"
                  column="span.event_error_count"
                  op=">"
                  value="0"
                  tooltip="Filter spans with errors"
                  @click="menu = false"
                />
              </v-col>
            </v-row>

            <v-divider class="my-6" />

            <v-row>
              <v-col class="grey--text text--darken-3"
                >For example, <strong>where span.duration > 100ms</strong>.</v-col
              >
            </v-row>

            <v-row class="mb-n3">
              <v-col cols="auto" class="pt-4 pr-2">
                <v-avatar color="grey" size="30">
                  <span class="white--text text-h6">1</span>
                </v-avatar>
              </v-col>
              <v-col>
                <v-autocomplete
                  v-model="form.column"
                  :loading="columnsDs.loading"
                  :items="columnsDs.items"
                  :rules="form.rules.column"
                  label="Column"
                  return-object
                  outlined
                  dense
                  clearable
                  hide-details="auto"
                ></v-autocomplete>
              </v-col>
            </v-row>

            <v-row class="mb-n3">
              <v-col cols="auto" class="pt-4 pr-2">
                <v-avatar color="grey" size="30">
                  <span class="white--text text-h6">2</span>
                </v-avatar>
              </v-col>
              <v-col>
                <v-autocomplete
                  v-model="form.op"
                  :rules="form.rules.op"
                  label="Operator"
                  :items="opItems"
                  outlined
                  dense
                  clearable
                  hide-details="auto"
                  class="fit"
                  style="width: 200px"
                ></v-autocomplete>
              </v-col>
            </v-row>

            <v-row>
              <v-col cols="auto" class="pt-4 pr-2">
                <v-avatar color="grey" size="30">
                  <span class="white--text text-h6">3</span>
                </v-avatar>
              </v-col>
              <v-col>
                <Combobox
                  v-model="form.columnValue"
                  :data-source="valuesDs"
                  :rules="form.rules.columnValue"
                  label="Value"
                  :placeholder="valuePlaceholder"
                  :hint="valueHint"
                  :disabled="valueDisabled"
                  dense
                />
              </v-col>
            </v-row>

            <v-row>
              <v-spacer />
              <v-col cols="auto">
                <v-btn :disabled="!form.isValid" class="primary" @click="addFilter">Filter</v-btn>
              </v-col>
            </v-row>
          </v-form>
        </v-col>
      </v-row>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { useDataSource, Item } from '@/use/datasource'
import { UseUql } from '@/use/uql'

// Components
import Combobox from '@/components/Combobox.vue'
import UqlChip from '@/components/UqlChip.vue'

// Misc
import { requiredRule } from '@/util/validation'
import { quote } from '@/util/string'
import { AttrKey } from '@/models/otel'

interface Column extends Item {
  ordered?: boolean
  searchable?: boolean
}

enum Op {
  Exists = 'exists',
  NotExists = 'not exists',

  Like = 'like',
  NotLike = 'not like',

  Contains = 'contains',
  NotContains = 'not contains',
}

export default defineComponent({
  name: 'WhereFilterMenu',
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
  },

  setup(props) {
    const route = useRoute()
    const menu = shallowRef(false)
    const formRef = shallowRef()
    const form = useForm()

    const columnsDs = useDataSource<Column>(
      () => {
        if (!menu.value) {
          return null
        }

        const { projectId } = route.value.params
        return {
          url: `/internal/v1/tracing/${projectId}/attributes?with_columns`,
          params: props.axiosParams,
        }
      },
      { suggestSearchInput: true },
    )

    const opItems = computed((): string[] => {
      const ops = []
      ops.push(Op.Contains, Op.NotContains, Op.Like, Op.NotLike)
      ops.push('=', '!=', '<', '<=', '>', '>=')
      ops.push(Op.Exists, Op.NotExists)
      return ops
    })

    const valuesDs = useDataSource(
      () => {
        if (!menu.value || !form.column || !form.column.value) {
          return
        }

        const { projectId } = route.value.params
        return {
          url: `/internal/v1/tracing/${projectId}/attributes/${form.column.value}`,
          params: {
            ...props.axiosParams,
          },
        }
      },
      { suggestSearchInput: true },
    )

    const valuePlaceholder = computed((): string => {
      switch (form.op) {
        case Op.Like:
        case Op.NotLike:
          return '%substring% or %suffix or prefix%'
        case Op.Contains:
        case Op.NotContains:
          return 'substr1|substr2|substr3'
        default:
          return ''
      }
    })

    const valueHint = computed((): string => {
      switch (form.op) {
        case Op.Like:
        case Op.NotLike:
          return '"%" matches zero or more characters'
        case Op.Contains:
        case Op.NotContains:
          return 'Case-insensitive options separated with "|"'
        default:
          return ''
      }
    })

    const valueDisabled = computed((): boolean => {
      switch (form.op) {
        case Op.Exists:
        case Op.NotExists:
          return true
        default:
          return false
      }
    })

    function addFilter() {
      setTimeout(() => {
        if (!form.column || !form.op) {
          return
        }

        const editor = props.uql.createEditor()

        if (valueDisabled.value) {
          editor.add(`where ${form.column.value} ${form.op}`)
        } else {
          const value = form.columnValue?.value ?? ''
          editor.add(`where ${form.column.value} ${form.op} ${quote(value)}`)
        }

        props.uql.commitEdits(editor)

        form.column = undefined
        form.op = ''
        form.columnValue = undefined
        formRef.value.resetValidation()

        menu.value = false
      }, 10)
    }

    watch(valueDisabled, (disabled) => {
      if (disabled) {
        form.columnValue = undefined
      }
    })

    return {
      AttrKey,
      menu,
      form,
      formRef,

      columnsDs,
      opItems,
      valuesDs,
      valuePlaceholder,
      valueHint,
      valueDisabled,

      addFilter,
    }
  },
})

function useForm() {
  const isValid = shallowRef(false)
  const rules = {
    column: [requiredRule],
    op: [requiredRule],
    columnValue: [],
  }

  const column = shallowRef<Column>()
  const op = shallowRef('')
  const columnValue = shallowRef<Item>()

  return proxyRefs({
    isValid,
    rules,

    column,
    op,
    columnValue,
  })
}
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
