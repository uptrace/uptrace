<template>
  <v-card flat :max-height="maxHeight">
    <v-list dense>
      <template v-for="item in tree">
        <v-menu
          v-if="item.children"
          :key="item.name"
          open-on-hover
          offset-x
          transition="slide-x-transition"
        >
          <template #activator="{ on, attrs }">
            <v-list-item v-bind="attrs" v-on="on" @click.stop>
              <v-list-item-content>
                <v-list-item-title>{{ item.name }}</v-list-item-title>
              </v-list-item-content>
              <v-list-item-icon class="align-self-center">
                <v-icon>mdi-menu-right</v-icon>
              </v-list-item-icon>
            </v-list-item>
          </template>

          <DashTree :tree="item.children" @change="$emit('change', $event)" />
        </v-menu>

        <v-list-item
          v-else
          :key="item.id"
          :to="{ params: { dashId: item.id } }"
          exact
          @click="$emit('change')"
        >
          <v-list-item-content>
            <v-list-item-title>{{ item.name }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </template>
    </v-list>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { DashTree } from '@/metrics/use-dashboards'

export default defineComponent({
  name: 'DashTree',

  props: {
    tree: {
      type: Array as PropType<DashTree[]>,
      required: true,
    },
    maxHeight: {
      type: Number,
      default: 420,
    },
  },
})
</script>

<style lang="scss" scoped></style>
