<template>
  <tr>
    <td>
      <v-avatar size="12" :color="annotation.color" class="mr-2" />
      {{ annotation.name }}
    </td>
    <td>
      <AnnotationAttrs :attrs="annotation.attrs" small />
    </td>
    <td class="text-no-wrap text-center">
      <v-btn
        :to="{ name: 'AnnotationShow', params: { annotationId: annotation.id } }"
        icon
        title="Edit annotation"
        ><v-icon>mdi-pencil-outline</v-icon></v-btn
      >
      <v-btn
        :loading="annotationMan.pending"
        icon
        title="Delete annotation"
        @click="deleteAnnotation"
        ><v-icon>mdi-delete-outline</v-icon></v-btn
      >
    </td>
    <td class="text-right text-no-wrap">
      <DateValue :value="annotation.createdAt" format="relative" />
    </td>
  </tr>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'
import { useAnnotationManager, Annotation } from '@/org/use-annotations'

// Components
import AnnotationAttrs from '@/alerting/AnnotationAttrs.vue'

export default defineComponent({
  name: 'AnnotationsTableRow',
  components: { AnnotationAttrs },

  props: {
    annotation: {
      type: Object as PropType<Annotation>,
      required: true,
    },
  },

  setup(props, ctx) {
    const confirm = useConfirm()
    const annotationMan = useAnnotationManager()

    function deleteAnnotation() {
      confirm
        .open(
          'Delete annotation',
          `Do you really want to delete "${props.annotation.name}" annotation?`,
        )
        .then(() => annotationMan.del(props.annotation))
        .then((annotation) => ctx.emit('change', annotation))
        .catch(() => {})
    }

    return {
      annotationMan,

      deleteAnnotation,
    }
  },
})
</script>

<style lang="scss" scoped></style>
