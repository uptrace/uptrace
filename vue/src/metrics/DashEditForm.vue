<template>
  <v-form v-model="isValid" @submit.prevent="update">
    <v-card>
      <v-toolbar color="light-blue lighten-5" flat dense>
        <v-toolbar-title>Edit dashboard</v-toolbar-title>
      </v-toolbar>

      <div class="py-4 px-6">
        <v-row class="mb-n2">
          <v-col>
            <v-text-field
              v-model="name"
              label="Dashboard name"
              :rules="rules.name"
              dense
              filled
              background-color="grey lighten-4"
              required
              autofocus
            />
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <slot name="prepend-actions" />
            <v-btn type="submit" color="primary" :disabled="!isValid" :loading="dashMan.pending"
              >Update</v-btn
            >
          </v-col>
        </v-row>
      </div>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Composables
import { useDashManager, Dashboard } from '@/metrics/use-dashboards'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashEditForm',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const name = shallowRef('')
    const isValid = shallowRef(false)
    const rules = {
      name: [requiredRule],
    }

    const dashMan = useDashManager()

    watch(
      () => props.dashboard.name,
      (dashName) => {
        name.value = dashName
      },
      { immediate: true },
    )

    function update() {
      if (!isValid.value) {
        return
      }

      dashMan.update({ name: name.value }).then((dash) => {
        ctx.emit('update', dash)
      })
    }

    return {
      name,
      isValid,
      rules,

      dashMan,
      update,
    }
  },
})
</script>

<style lang="scss" scoped></style>
