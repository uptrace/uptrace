<template>
  <div style="position: relative">
    <v-progress-linear
      v-if="values.loading"
      absolute
      indeterminate
      class="mx-4"
    ></v-progress-linear>
    <div class="pa-3">
      <v-row no-gutters align="center">
        <v-col :cols="hasSearch ? 'auto' : 12">
          <div class="cursor-pointer" @click="expandedInternal = !expandedInternal">
            <v-icon class="mr-1">{{
              expandedInternal ? 'mdi-chevron-down' : 'mdi-chevron-up'
            }}</v-icon>
            <span class="text-subtitle-2">{{ attr }}</span>
          </div>
        </v-col>
        <v-col class="pl-4">
          <v-text-field
            v-if="hasSearch"
            v-model="values.searchQuery"
            :loading="values.loading"
            prepend-inner-icon="mdi-magnify"
            outlined
            dense
            hide-details="auto"
            clearable
          ></v-text-field>
        </v-col>
      </v-row>

      <v-row v-if="values.searchQuery" no-gutters class="mt-1">
        <v-col>
          <v-list dense>
            <v-list-item @click="$emit('click:add-query', likeQuery)">
              <v-list-item-icon class="mr-4">
                <v-icon>mdi-magnify</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ likeQuery }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="$emit('click:add-query', notLikeQuery)">
              <v-list-item-icon class="mr-4">
                <v-icon>mdi-magnify</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ notLikeQuery }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>

      <v-row v-if="expandedInternal" no-gutters class="mt-1">
        <v-col>
          <SpanFacetBody
            :value="value"
            :items="values.items"
            :search-query.sync="values.searchQuery"
            show-search
            @input="$emit('input', $event)"
            @click:close="$emit('click:close')"
          >
          </SpanFacetBody>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useSpanAttrValues } from '@/tracing/query/use-span-facets'

// Components
import SpanFacetBody from '@/tracing/query/SpanFacetBody.vue'

// Utilities
import { quote } from '@/util/string'

export default defineComponent({
  name: 'SpanFacet',
  components: { SpanFacetBody },

  props: {
    axiosParams: {
      type: undefined as unknown as PropType<Record<string, any> | null>,
      required: true,
    },
    value: {
      type: Array as PropType<string[]>,
      default: undefined,
    },
    attr: {
      type: String,
      required: true,
    },
    expanded: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const expandedInternal = shallowRef(props.expanded)

    const values = useSpanAttrValues(() => {
      if (!props.axiosParams) {
        return props.axiosParams
      }
      if (!expandedInternal.value) {
        return { _: undefined }
      }
      return {
        ...props.axiosParams,
        attr_key: props.attr,
      }
    })

    const hasSearch = computed(() => {
      return expandedInternal.value && (values.items.length > 10 || values.searchQuery)
    })

    const likeQuery = computed(() => {
      const value = `%${values.searchQuery}%`
      return `${props.attr} like ${quote(value)}`
    })

    const notLikeQuery = computed(() => {
      const value = `%${values.searchQuery}%`
      return `${props.attr} not like ${quote(value)}`
    })

    return { expandedInternal, values, hasSearch, likeQuery, notLikeQuery }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}

.v-input {
  & ::v-deep .v-input__slot {
    min-height: 32px !important;
    height: 32px !important;
  }

  & ::v-deep .v-input__prepend-inner,
  & ::v-deep .v-input__append-inner {
    margin-top: 5px !important;
  }
}
</style>
