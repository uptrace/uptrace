<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        Group by
      </v-btn>
    </template>
    <v-form ref="form" v-model="isValid">
      <v-card width="400px">
        <v-card-text class="py-6">
          <v-row>
            <v-col class="space-around no-transform">
              <UqlChip
                v-for="column in groupColumns"
                :key="column"
                :uql="uql"
                :group="column"
                @click="menu = false"
              />
            </v-col>
          </v-row>

          <div class="mt-2 mb-3 d-flex align-center">
            <v-divider />
            <div class="mx-2 grey--text text--lighten-1">or</div>
            <v-divider />
          </div>

          <v-row class="mb-n1">
            <v-col>Select a grouping column.</v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <SimpleSuggestions
                v-model="column"
                :loading="suggestions.loading"
                :suggestions="suggestions"
                :rules="rules.column"
                label="Column"
                dense
                class="fit"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-spacer />
            <v-col cols="auto">
              <v-btn :disabled="!isValid" class="mr-2 secondary" @click="add">Add</v-btn>
              <v-btn :disabled="!isValid" class="primary" @click="replace">Replace</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { useSuggestions, Suggestion } from '@/use/suggestions'
import { UseUql } from '@/use/uql'

// Components
import SimpleSuggestions from '@/components/SimpleSuggestions.vue'
import UqlChip from '@/components/UqlChip.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { requiredRule } from '@/util/validation'

const groupColumns = [
  AttrKey.spanGroupId,
  AttrKey.serviceName,
  AttrKey.hostName,
  AttrKey.dbOperation,
]

export default defineComponent({
  name: 'GroupByMenu',
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
    const column = shallowRef<Suggestion>()

    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = {
      column: [requiredRule],
    }

    const suggestions = useSuggestions(
      () => {
        if (!menu.value) {
          return null
        }

        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/suggestions/attributes`,
          params: {
            ...props.axiosParams,
          },
        }
      },
      { suggestSearchInput: true },
    )

    function add() {
      updateQuery(false)
    }

    function replace() {
      updateQuery(true)
    }

    function updateQuery(replace = false) {
      if (!column.value) {
        return
      }

      const editor = props.uql.createEditor()
      if (replace) {
        editor.replaceGroupBy(column.value.text)
      } else {
        editor.addGroupBy(column.value.text)
      }
      props.uql.commitEdits(editor)

      column.value = undefined
      form.value.resetValidation()
      menu.value = false
    }

    return {
      menu,

      form,
      isValid,
      rules,
      suggestions,

      groupColumns,
      column,

      add,
      replace,
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
