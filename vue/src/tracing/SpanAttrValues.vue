<template>
  <PagedGroupsCard
    :date-range="dateRange"
    :systems="[system]"
    :loading="groups.loading"
    :groups="groups.items"
    :columns="groups.columns"
    :plottable-columns="groups.plottableColumns"
    :plotted-columns="plottedColumns"
    :order="groups.order"
    :axios-params="groups.axiosParams"
    show-plotted-column-items
    hide-actions
  >
    <template v-if="queryPartItems.length" #actions>
      <v-slide-group v-model="activeQueryParts" multiple center-active show-arrows>
        <v-slide-item
          v-for="(item, i) in queryPartItems"
          :key="item.value"
          v-slot="{ active, toggle }"
          :value="item.value"
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
            {{ item.text }}
          </v-btn>
        </v-slide-item>
      </v-slide-group>
    </template>
  </PagedGroupsCard>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { createQueryEditor } from '@/use/uql'
import { UseDateRange } from '@/use/date-range'
import { injectQueryStore } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'

// Utilities
import { isSpanSystem, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpanAttrValues',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      required: true,
    },
    attrKey: {
      type: String,
      required: true,
    },
    attrValue: {
      type: undefined,
      required: true,
    },
  },

  setup(props) {
    const plottedColumns = [AttrKey.spanCountPerMin, `p50(${AttrKey.spanDuration})`]
    const activeQueryParts = shallowRef(plottedColumns)

    const queryPartItems = computed(() => {
      const items = [
        {
          value: AttrKey.spanCount,
        },
        {
          value: AttrKey.spanCountPerMin,
        },
        {
          value: `min(${AttrKey.spanTime})`,
        },
        {
          value: `max(${AttrKey.spanTime})`,
        },
      ]

      if (isSpanSystem(props.system)) {
        items.push(
          {
            value: `p50(${AttrKey.spanDuration})`,
          },
          {
            value: `p90(${AttrKey.spanDuration})`,
          },
          {
            value: `p99(${AttrKey.spanDuration})`,
          },
        )
      }

      return items.map((item) => {
        return {
          value: item.value,
          text: item.value,
        }
      })
    })

    const { where } = injectQueryStore()

    const query = computed(() => {
      const editor = createQueryEditor().add(`group by ${props.attrKey}`)

      for (let item of queryPartItems.value) {
        if (activeQueryParts.value.includes(item.value)) {
          editor.add(item.value)
        }
      }

      editor.where(props.attrKey, 'exists').add(where.value)

      return editor.toString()
    })

    const groups = useGroups(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.system,
        group_id: props.groupId,
        query: query.value,
      }
    })

    return {
      plottedColumns,

      query,
      groups,

      activeQueryParts,
      queryPartItems,
    }
  },
})
</script>

<style lang="scss" scoped></style>
