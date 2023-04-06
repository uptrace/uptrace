<template>
  <div>
    <v-simple-table class="v-data-table--border table-scroll-target">
      <colgroup>
        <col />
        <col />
      </colgroup>

      <thead class="v-data-table-header">
        <tr>
          <th>Key</th>
          <th class="target">Value</th>
        </tr>
      </thead>

      <tbody v-if="!attrKeys.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99">There are no attributes matching the filters.</td>
        </tr>
      </tbody>

      <tbody>
        <tr v-for="key in attrKeys" :key="key">
          <th>{{ key }}</th>
          <td>
            <CodeOrText :value="attrs[key]" :name="key">
              <template #text>
                <KeyValueFilterLink
                  v-if="groupId"
                  :date-range="dateRange"
                  :name="key"
                  :value="attrs[key]"
                  :project-id="$route.params.projectId"
                  :system="system"
                  :group-id="groupId"
                  :is-event="isEvent"
                />
              </template>
            </CodeOrText>
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'
import { AttrMap } from '@/models/span'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import CodeOrText from '@/components/CodeOrText.vue'
import KeyValueFilterLink from '@/tracing/KeyValueFilterLink.vue'

export default defineComponent({
  name: 'AttrsTable',
  components: { CodeOrText, KeyValueFilterLink },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
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
    attrs: {
      type: Object as PropType<AttrMap>,
      required: true,
    },
  },

  setup(props) {
    const attrKeys = computed((): string[] => {
      const keys = Object.keys(props.attrs)
      keys.sort()
      return keys
    })

    const isEvent = computed((): boolean => {
      return isEventSystem(props.system)
    })

    return {
      AttrKey,
      attrKeys,
      isEvent,
    }
  },
})
</script>

<style lang="scss" scoped></style>
