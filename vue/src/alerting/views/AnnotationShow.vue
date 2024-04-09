<template>
  <v-container fluid class="fill-height bg--none">
    <v-row align="center" justify="center">
      <v-col cols="auto">
        <v-skeleton-loader v-if="!annotation.data" width="600" type="card"></v-skeleton-loader>

        <v-card v-else width="600">
          <v-toolbar flat color="bg--none-primary">
            <v-breadcrumbs :items="breadcrumbs" divider=">" large class="pl-0"></v-breadcrumbs>
          </v-toolbar>

          <div class="pa-4">
            <AnnotationForm
              :annotation="annotation.data"
              @click:close="$router.push({ name: 'AnnotationList' })"
            />
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useProject } from '@/org/use-projects'
import { useTitle } from '@vueuse/core'
import { useAnnotation } from '@/org/use-annotations'

// Components
import AnnotationForm from '@/alerting/AnnotationForm.vue'

export default defineComponent({
  name: 'AnnotationShow',
  components: { AnnotationForm },

  setup() {
    useTitle('Annotation')

    const project = useProject()
    const annotation = useAnnotation()

    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: project.data?.name ?? 'Project',
        to: {
          name: 'ProjectShow',
        },
        exact: true,
      })

      bs.push({
        text: 'Annotations',
        to: {
          name: 'AnnotationList',
        },
        exact: true,
      })

      bs.push({ text: 'Annotation' })

      return bs
    })

    return {
      annotation,
      breadcrumbs,
    }
  },
})
</script>

<style lang="scss" scoped></style>
