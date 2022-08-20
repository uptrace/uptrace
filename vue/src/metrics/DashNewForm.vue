<template>
  <v-form v-model="isValid" @submit.prevent="create">
    <v-card>
      <v-toolbar color="light-blue lighten-5" flat dense>
        <v-toolbar-title>New dashboard</v-toolbar-title>
      </v-toolbar>

      <v-card-text>
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
              >Create</v-btn
            >
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useDashManager } from '@/metrics/use-dashboards'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashNewForm',

  setup(props, ctx) {
    const { route } = useRouter()

    const name = shallowRef('')
    const isValid = shallowRef(false)
    const rules = {
      name: [requiredRule],
    }

    const dashMan = useDashManager()

    function create() {
      if (!isValid.value) {
        return
      }

      const { projectId } = route.value.params
      dashMan
        .create({
          name: name.value,
          projectId: parseInt(projectId, 10),
        })
        .then((dash) => {
          name.value = ''
          ctx.emit('create', dash)
        })
    }

    return {
      name,
      isValid,
      rules,

      dashMan,
      create,
    }
  },
})
</script>

<style lang="scss" scoped></style>
