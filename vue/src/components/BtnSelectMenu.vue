<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on, attrs }">
      <v-btn
        :disabled="disabled"
        :color="color"
        :elevation="elevation"
        :outlined="outlined"
        class="text-none"
        :class="targetClass"
        v-bind="attrs"
        v-on="on"
      >
        <span v-if="active">{{ active.short ? active.short : active.text }}</span>
        <span v-else>Please select</span>
        <v-icon>mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <v-card>
      <v-list dense>
        <v-list-item-group :value="value" mandatory>
          <v-list-item
            v-for="item in items"
            :key="item.value"
            :value="item.value"
            @click="$emit('input', item.value)"
          >
            <slot name="item" :item="item">
              <v-list-item-content>
                <v-list-item-title>{{ item.text }}</v-list-item-title>
                <v-list-item-subtitle v-if="item.hint">{{ item.hint }}</v-list-item-subtitle>
              </v-list-item-content>
            </slot>
          </v-list-item>
        </v-list-item-group>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

interface Item {
  value: string
  text: string
  short?: string
}

export default defineComponent({
  name: 'BtnSelectMenu',

  props: {
    value: {
      type: [Number, String],
      default: undefined,
    },
    items: {
      type: Array as PropType<Item[]>,
      required: true,
    },
    dense: {
      type: Boolean,
      default: false,
    },
    color: {
      type: String,
      default: undefined,
    },
    elevation: {
      type: Number,
      default: 0,
    },
    outlined: {
      type: Boolean,
      default: false,
    },
    targetClass: {
      type: String,
      default: '',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)

    const active = computed(() => {
      return props.items.find((item) => item.value === props.value)
    })

    watch(
      active,
      (active) => {
        if (active === undefined && props.items.length) {
          ctx.emit('input', props.items[0].value)
        }
      },
      { immediate: true },
    )

    return { menu, active }
  },
})
</script>

<style lang="scss" scoped></style>
