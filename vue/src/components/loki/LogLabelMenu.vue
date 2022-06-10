<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>{{ label }}</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <XList
      :loading="labelValues.loading"
      :items="labelValues.items"
      return-object
      @input="addFilter('=', $event.text)"
    >
      <template #item="{ item }">
        <v-list-item-content>
          <v-list-item-title>
            {{ truncate(item.text, { length: 60 }) }}
          </v-list-item-title>
        </v-list-item-content>

        <v-list-item-action class="my-0" @click.stop="addFilter('!=', item.text)">
          <v-btn icon>
            <v-icon small>mdi-not-equal</v-icon>
          </v-btn>
        </v-list-item-action>
      </template>
    </XList>
  </v-menu>
</template>
<script lang="ts">
import { truncate } from 'lodash'
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useLabelValues } from '@/components/loki/logql'

// Components
import XList from '@/components/XList.vue'

export default defineComponent({
  name: 'LogLabelMenu',
  components: { XList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
  },
  setup(props, ctx) {
    const { route } = useRouter()
    const menu = shallowRef(false)

    const labelValues = useLabelValues(() => {
      if (!menu.value) {
        return null
      }

      const { projectId } = route.value.params
      return {
        url: `/${projectId}/loki/api/v1/label/${props.label}/values`,
        params: {
          ...props.dateRange.lokiParams(),
        },
      }
    })

    function addFilter(op: string, value: string) {
      ctx.emit('click', { op, value })
      menu.value = false
    }

    return { menu, labelValues, addFilter, truncate }
  },
})
</script>

<style lang="scss" scoped></style>
