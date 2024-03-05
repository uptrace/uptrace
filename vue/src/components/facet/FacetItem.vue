<template>
  <div style="position: relative">
    <v-progress-linear
      v-if="values.loading"
      absolute
      indeterminate
      class="mx-4"
    ></v-progress-linear>

    <div class="pa-3">
      <v-row dense align="center">
        <v-col>
          <div class="cursor-pointer" @click="expandedInternal = !expandedInternal">
            <v-icon class="mr-1">{{
              expandedInternal ? 'mdi-chevron-down' : 'mdi-chevron-up'
            }}</v-icon>
            <span class="text-subtitle-2">{{ attr }}</span>
            <v-btn
              v-if="pinned"
              :loading="pending"
              icon
              small
              title="Unpin attribute"
              class="ml-1"
              @click.stop.prevent="$emit('click:unpin')"
            >
              <v-icon size="20" color="green darken-1">mdi-pin</v-icon>
            </v-btn>
            <v-btn
              v-else
              :loading="pending"
              icon
              small
              title="Pin attribute to the top"
              class="ml-1"
              @click.stop.prevent="$emit('click:pin')"
            >
              <v-icon size="20">mdi-pin-outline</v-icon>
            </v-btn>
          </div>
        </v-col>
      </v-row>

      <v-row v-if="showSearch" dense>
        <v-col>
          <v-text-field
            v-if="showSearch"
            v-model="values.searchInput"
            :loading="values.loading"
            prepend-inner-icon="mdi-magnify"
            outlined
            dense
            hide-details="auto"
            clearable
          ></v-text-field>
        </v-col>
      </v-row>

      <v-row v-if="values.searchInput" dense class="mt-1">
        <v-col>
          <v-list dense>
            <v-list-item class="px-1" @click="$emit('update:filter', likeFilter)">
              <v-list-item-icon class="mr-4">
                <v-icon>mdi-magnify</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ filterString(likeFilter) }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item class="px-1" @click="$emit('update:filter', notLikeFilter)">
              <v-list-item-icon class="mr-4">
                <v-icon>mdi-magnify</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>{{ filterString(notLikeFilter) }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>

      <v-row v-if="expandedInternal" dense>
        <v-col>
          <FacetItemBody
            :value="value"
            :items="values.filteredItems"
            :search-query.sync="values.searchInput"
            @input="$emit('input', $event)"
            @click:close="$emit('click:close')"
          >
          </FacetItemBody>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { Filter } from '@/components/facet/types'

// Components
import FacetItemBody from '@/components/facet/FacetItemBody.vue'

export default defineComponent({
  name: 'SpanFacet',
  components: { FacetItemBody },

  props: {
    component: {
      type: String,
      required: true,
    },
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
    pinned: {
      type: Boolean,
      required: true,
    },
    pending: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()
    const expandedInternal = shallowRef(props.expanded)

    const values = useDataSource(() => {
      if (!props.axiosParams) {
        return props.axiosParams
      }
      if (!expandedInternal.value) {
        return undefined
      }
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/${props.component}/${projectId}/attributes/${props.attr}`,
        params: {
          ...props.axiosParams,
        },
        debounce: 500,
      }
    })

    const showSearch = computed(() => {
      return expandedInternal.value && (values.items.length > 10 || values.searchInput)
    })

    const likeFilter = computed(() => {
      return {
        attr: props.attr,
        op: 'like',
        value: [`%${values.searchInput}%`],
      }
    })

    const notLikeFilter = computed(() => {
      return {
        attr: props.attr,
        op: 'not like',
        value: [`%${values.searchInput}%`],
      }
    })

    function filterString(f: Filter) {
      return `${f.attr} ${f.op} ${f.value}`
    }

    return {
      expandedInternal,
      values,
      showSearch,

      likeFilter,
      notLikeFilter,
      filterString,
    }
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
