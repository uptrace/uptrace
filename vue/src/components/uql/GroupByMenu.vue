<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        <span>Group by</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-form ref="form" v-model="isValid" @submit.prevent="addFilter">
      <v-card width="400px">
        <v-card-text class="py-6">
          <v-row>
            <v-col class="space-around no-transform">
              <UqlChip
                v-for="columnUql in groupColumns"
                :key="columnUql"
                :uql="uql"
                :group="columnUql"
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
              <v-btn type="submit" :disabled="!isValid" class="primary">Group by</v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-form>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

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

const groupColumns = [xkey.spanGroupId, xkey.serviceName, xkey.hostName, xkey.dbOperation]

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
          url: `/api/tracing/${projectId}/suggestions/attributes`,
          params: props.axiosParams,
        }
      },
      { suggestSearchInput: true },
    )

    function addFilter() {
      if (!column.value) {
        return
      }

      fastGroupBy(column.value.text)

      column.value = undefined
      form.value.resetValidation()
    }

    function fastGroupBy(column: string) {
      const editor = props.uql.createEditor()
      editor.replaceGroupBy(column)
      props.uql.commitEdits(editor)

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

      addFilter,
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

.no-transform :deep(.v-btn) {
  padding: 0 12px !important;
  text-transform: none;
}

.space-around :deep(.v-chip) {
  margin: 4px;
}
</style>
