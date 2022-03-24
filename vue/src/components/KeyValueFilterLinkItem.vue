<template>
  <div>
    <v-menu v-if="menuItems.length" v-model="menu" offset-y>
      <template #activator="{ on }">
        <v-btn icon v-on="on"><v-icon>mdi-dots-vertical</v-icon></v-btn>
      </template>

      <v-list dense>
        <v-list-item v-for="item in menuItems" :key="item.title" v-bind="item.link">
          <v-list-item-content class="text-body-2">
            <v-list-item-title v-text="item.title"></v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>

    <XText :value="value" :name="name" />
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { buildWhere, buildGroupBy } from '@/use/uql'

// Utilities
import { xkey } from '@/models/otelattr'
import { truncateMiddle } from '@/util/string'
import { createFormatter, unitFromName } from '@/util/fmt'

interface MenuItem {
  title: string
  link: Record<string, any>
}

export default defineComponent({
  name: 'KeyValueFilterLinkItem',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    projectId: {
      type: [Number, String],
      default: undefined,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      default: undefined,
    },
    name: {
      type: String,
      required: true,
    },
    value: {
      type: undefined,
      required: true,
    },
    filterable: {
      type: Boolean,
      default: true,
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const { route } = useRouter()

    const query = computed(() => {
      if (route.value.query.query) {
        return route.value.query.query
      }
      const query = buildGroupBy(xkey.spanGroupId)
      if (props.groupId) {
        return query + ` | where ${xkey.spanGroupId} = ${props.groupId}`
      }
      return query
    })

    const menuItems = computed((): MenuItem[] => {
      if (!props.filterable) {
        return []
      }

      const items = [
        {
          title: `Group by ${props.name}`,
          link: link({ query: buildGroupBy(props.name) }),
        },
      ]

      const ops = ['=', '!=']
      if (typeof props.value === 'number') {
        ops.push('<', '<=', '>', '>=')
      }

      for (let op of ops) {
        items.push({
          title: `${props.name} ${op} ${format(props.value)}`,
          link: link({
            query: [query.value, buildWhere(props.name, op, props.value as any)].join(' | '),
          }),
        })
      }

      for (let op of ['exists', 'not exists']) {
        items.push({
          title: `${props.name} ${op}`,
          link: link({
            query: [query.value, buildWhere(props.name, op)].join(' | '),
          }),
        })
      }

      return items
    })

    function format(v: any): string {
      const fmt = createFormatter(unitFromName(props.name))
      return truncateMiddle(fmt(v))
    }

    function link(query: Record<string, any> = {}) {
      query = {
        ...query,
        ...props.dateRange.queryParams(),
        system: props.system,
      }

      return {
        to: {
          name: route.value.name === 'SpanList' ? 'SpanList' : 'GroupList',
          query,
        },
        exact: true,
      }
    }

    return { menu, menuItems }
  },
})
</script>

<style lang="scss" scoped></style>
